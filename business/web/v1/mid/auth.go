package mid

import (
	"context"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/web/auth"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"github.com/google/uuid"
	"net/http"
)

// Authenticate validates a JWT from the `Authorization` header.
func Authenticate(a *auth.Auth) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			claims, err := a.Authenticate(ctx, r.Header.Get("authorization"))
			if err != nil {
				return auth.NewAuthError("authenticate: failed: %s", err)
			}

			/* Put the claims into the ctx just in case the handler func may need it for any reason. Note: The business layer should never ever
			need this, we're talking about the application layer needing it*/
			ctx = auth.SetClaims(ctx, claims)

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}

// Authorize executes the specified role and does not extract any domain data.
/* Authorize validates that an authenticated user has at least one role from a specified list. This method constructs the actual function
that is used.*/
func Authorize(a *auth.Auth, rule string) web.Middleware {
	m := func(handler web.Handler) web.Handler {
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			// get the claims out of ctx. Here we use the claims that we retrieved from jwt in the authentication step:
			claims := auth.GetClaims(ctx)
			if claims.Subject == "" {
				return auth.NewAuthError("authorize: you are not authorized for that action, no claims")
			}

			if err := a.Authorize(ctx, claims, uuid.UUID{}, rule); err != nil {
				return auth.NewAuthError("authorize: you are not authorized for that action, claims[%v] rule[%v]: %s", claims.Roles, rule, err)
			}

			return handler(ctx, w, r)
		}

		return h
	}

	return m
}
