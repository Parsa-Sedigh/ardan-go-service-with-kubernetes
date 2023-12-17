package testgrp

import (
	"context"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/foundation/web"
	"net/http"
)

// Test is our example route
func Test(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// validate the data
	// call into the business layer
	// return errors or handle OK response

	status := struct {
		Status string
	}{
		Status: "ok",
	}

	return web.Respond(ctx, w, status, http.StatusOK)
}
