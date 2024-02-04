package v1

import (
	"net/http"
)

// UserTests holds methods for each user subtest. This type allows passing dependencies for tests while still providing a
// convenient syntax when subtests are registered.
type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}
