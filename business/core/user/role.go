package user

import "fmt"

// Set of possible roles for a user.
var (
	RoleAdmin = Role{"ADMIN"}
	RoleUser  = Role{"USER"}
)

// Set of known roles.
var roles = map[string]Role{
	RoleAdmin.name: RoleAdmin,
	RoleUser.name:  RoleUser,
}

// Role represents a role in the system.
type Role struct {
	name string
}

// ParseRole parses the string value and returns a role if one exists.
// This is a validation func. It should only be called at the app layer. The business layer accepts values of the return type of `Parse`
// functions like this one and the app layer is accepting strings and validate that those strings are a valid set(enum).
func ParseRole(value string) (Role, error) {
	role, exists := roles[value]
	if !exists {
		return Role{}, fmt.Errorf("invalid role %q", value)
	}

	return role, nil
}

// MustParseRole parses the string value and returns a role if one exists. If
// an error occurs the function panics.
// This function exists with the Parse func(like ParseRole), because when you're writing tests, you don't want to do the check, because otherwise
// it would make the test a bit longer. Nobody should be using `Must` funcs(like this one) in any application level code other than tests. That would be
// a code smell.
func MustParseRole(value string) Role {
	role, err := ParseRole(value)
	if err != nil {
		panic(err)
	}

	return role
}

// Name returns the name of the role.
func (r Role) Name() string {
	return r.name
}

// UnmarshalText implement the unmarshal interface for JSON conversions.
func (r *Role) UnmarshalText(data []byte) error {
	role, err := ParseRole(string(data))
	if err != nil {
		return err
	}

	r.name = role.name
	return nil
}

// MarshalText implement the marshal interface for JSON conversions.
func (r Role) MarshalText() ([]byte, error) {
	return []byte(r.name), nil
}

// Equal provides support for the go-cmp package and testing.
func (r Role) Equal(r2 Role) bool {
	return r.name == r2.name
}
