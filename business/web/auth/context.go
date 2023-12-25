package auth

import (
	"context"
)

// ctxKey represents the type of value for the context key.
/* The pattern is to define:
1. a key as a custom type unexported
2. a constant with an arbitrary value for this key and it's also unexported
3. define set and get funcs for setting and getting the value for this ctx key*/
type ctxKey int

// key is used to store/retrieve a Claims value from a context.Context.
const claimKey ctxKey = 1

// =============================================================================

// SetClaims stores the claims in the context.
func SetClaims(ctx context.Context, claims Claims) context.Context {
	return context.WithValue(ctx, claimKey, claims)
}

// GetClaims returns the claims from the context.
func GetClaims(ctx context.Context) Claims {
	v, ok := ctx.Value(claimKey).(Claims)
	if !ok {
		return Claims{}
	}
	return v
}
