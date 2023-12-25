package main

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/x509"
	"io"

	/* import the embed package although it's exported stuff are not used to run it's init funcs, because they are
	needed for go:embed */
	_ "embed"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/open-policy-agent/opa/rego"
)

// Core OPA policies.
var (
	// using go:embed, we don't need to do any file openings in code. So it's gonna become hard coded in the binary
	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)

func main() {
	privateKey, publicBlock, err := genKey()
	if err != nil {
		log.Fatal(err)
	}

	if err := genToken(privateKey, publicBlock); err != nil {
		log.Fatal(err)
	}
}

func genKey() (*rsa.PrivateKey, pem.Block, error) {
	//////////////////////////////// Generate a new private key. ////////////////////////////////
	//privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	//if err != nil {
	//	return nil, pem.Block{}, fmt.Errorf("generating key: %w", err)
	//}

	//// Create a file for the private key information in PEM form.
	//privateFile, err := os.Create("private.pem")
	//if err != nil {
	//	return nil, pem.Block{}, fmt.Errorf("creating private file: %w", err)
	//}
	//defer privateFile.Close()
	//
	//// Construct a PEM block for the private key.
	//privateBlock := pem.Block{
	//	/* You tend to not want to write in the file what type of private key this is. So we don't say it's "RSA private key", instead,
	//	we keep this label generic, so we just say the type is private key.*/
	//	Type:  "PRIVATE KEY",
	//	Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	//}
	//
	//// Write the private key to the private key file.
	//if err := pem.Encode(privateFile, &privateBlock); err != nil {
	//	return nil, pem.Block{}, fmt.Errorf("encoding to private file: %w", err)
	//}

	////////////////////////////////

	// read the private key from the file shared in the repo(DANGEROUS)
	file, err := os.Open("zarf/keys/54bb2165-71e1-41a6-af3e-7da4a0e1e2c1.pem")
	if err != nil {
		return nil, pem.Block{}, fmt.Errorf("opening key file: %w", err)
	}
	defer file.Close()

	pemData, err := io.ReadAll(io.LimitReader(file, 1024*1024))
	if err != nil {
		return nil, pem.Block{}, fmt.Errorf("reading auth private key: %w", err)
	}

	// convert the pem bytes into a private key
	// pk stands for private key
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(pemData)
	if err != nil {
		return nil, pem.Block{}, fmt.Errorf("parsing auth private key: %w", err)
	}

	// ==============================================================================

	// Create a file for the public key information in PEM form.
	publicFile, err := os.Create("public.pem")
	if err != nil {
		return nil, pem.Block{}, fmt.Errorf("creating public file: %w", err)
	}
	defer publicFile.Close()

	// Marshal the public key from the private key to PKIX.
	//asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	asn1Bytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, pem.Block{}, fmt.Errorf("marshaling public key: %w", err)
	}

	// Construct a PEM block for the public key.
	publicBlock := pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: asn1Bytes,
	}

	// Write the public key to the public key file.
	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return nil, pem.Block{}, fmt.Errorf("encoding to public file: %w", err)
	}

	fmt.Println("private and public key files generated")

	return privateKey, publicBlock, nil
}

func genToken(privateKey *rsa.PrivateKey, publicBlock pem.Block) error {
	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := struct {
		jwt.RegisteredClaims
		Roles []string
	}{
		RegisteredClaims: jwt.RegisteredClaims{
			// we can put userID as subject
			Subject:   "12345678",
			Issuer:    "service project",
			ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(8760 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		},
		Roles: []string{"ADMIN"},
	}

	// the signing method only needs to be constructed once
	method := jwt.GetSigningMethod(jwt.SigningMethodRS256.Name)

	// this token has the header and the payload, but it's not signed yet. We have to sign it after this line
	token := jwt.NewWithClaims(method, claims)
	token.Header["kid"] = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	// provide the private key to be used for the signature. str is the signed jwt.
	str, err := token.SignedString(privateKey)
	if err != nil {
		return fmt.Errorf("signing token: %w", err)
	}

	fmt.Println("************** TOKEN **************")
	fmt.Printf("%s\n\n", str)

	// ==============================================================================

	fmt.Println("************** PUBLIC KEY **************")

	if err := pem.Encode(os.Stdout, &publicBlock); err != nil {
		return fmt.Errorf("encoding to private file: %w", err)
	}

	fmt.Print("\n")

	// ==============================================================================

	/* validate the jwt in code. When tokens is sent back to us in a req, we can validate the jwt signature using the public key.
	In other words, we want to validate the jwt signature against our public key.*/
	parser := jwt.NewParser(jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Name}))

	var clm struct {
		jwt.RegisteredClaims
		Roles []string
	}

	// return the public key pair in this func
	kf := func(token *jwt.Token) (interface{}, error) {
		return &privateKey.PublicKey, nil
	}

	/* The ParseWithClaims func of the jwt api, doesn't know where the public key is. To give it the public key, we give it a func
	that knows how to get the public key.*/
	tkn, err := parser.ParseWithClaims(str, &clm, kf)
	if err != nil {
		return fmt.Errorf("parsing with claims: %w", err)
	}

	if !tkn.Valid {
		return errors.New("token not valid")
	}

	fmt.Println("TOKEN VALIDATED")

	/* Authentication completed, now we can start applying authorization. We got the claims of the jwt back to eventually do the
	authorization.*/

	ctx := context.Background()

	// ==============================================================================

	var b bytes.Buffer
	if err := pem.Encode(&b, &publicBlock); err != nil {
		return fmt.Errorf("OPA authentication failed: %w", err)
	}

	if err := opaPolicyEvaluationAuthentication(ctx, b.String(), str, clm.Issuer); err != nil {
		return fmt.Errorf("OPA authentication failed: %w", err)
	}

	fmt.Println("TOKEN VALIDATED BY OPA")

	// ==============================================================================

	if err := opaPolicyEvaluationAuthorization(ctx); err != nil {
		return fmt.Errorf("OPA authorization failed: %w", err)
	}

	fmt.Println("AUTH VALIDATED BY OPA")

	// ==============================================================================

	fmt.Printf("\n%#v\n", clm)

	return nil
}

// opaPolicyEvaluationAuthentication allows us to no longer write code in go to do the validation of jwt, instead we offload that work to
// opa using .rego files. So we don't need to rely on go-jwt APIs.
func opaPolicyEvaluationAuthentication(ctx context.Context, pem string, tokenString string, issuer string) error {
	const opaPackage = "ardan.rego"
	const rule = "auth"

	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthentication),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]any{
		"Key":   pem,
		"Token": tokenString,
		"ISS":   issuer,
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}

func opaPolicyEvaluationAuthorization(ctx context.Context) error {
	const rule = "rule_admin_only"
	const opaPackage string = "ardan.rego"

	query := fmt.Sprintf("x = data.%s.%s", opaPackage, rule)

	q, err := rego.New(
		rego.Query(query),
		rego.Module("policy.rego", opaAuthorization),
	).PrepareForEval(ctx)
	if err != nil {
		return err
	}

	input := map[string]any{
		"Roles":   []string{"ADMIN"},
		"Subject": "12345678",
		"UserID":  "12345678",
	}

	results, err := q.Eval(ctx, rego.EvalInput(input))
	if err != nil {
		return fmt.Errorf("query: %w", err)
	}

	if len(results) == 0 {
		return errors.New("no results")
	}

	result, ok := results[0].Bindings["x"].(bool)
	if !ok || !result {
		return fmt.Errorf("bindings results[%v] ok[%v]", results, ok)
	}

	return nil
}
