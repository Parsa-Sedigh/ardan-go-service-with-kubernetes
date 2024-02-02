package usergrp

import (
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/core/user"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/cview/user/summary"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/validate"
	"net/http"
	"net/mail"
	"time"

	"github.com/google/uuid"
)

// we have different filter functions depending on the route.

func parseFilter(r *http.Request) (user.QueryFilter, error) {
	const (
		filterByUserID           = "user_id"
		filterByEmail            = "email"
		filterByStartCreatedDate = "start_created_date"
		filterByEndCreatedDate   = "end_created_date"
		filterByName             = "name"
	)

	values := r.URL.Query()

	var filter user.QueryFilter

	if userID := values.Get(filterByUserID); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByUserID, err)
		}
		filter.WithUserID(id)
	}

	if email := values.Get(filterByEmail); email != "" {
		addr, err := mail.ParseAddress(email)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByEmail, err)
		}
		filter.WithEmail(*addr)
	}

	if createdDate := values.Get(filterByStartCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByStartCreatedDate, err)
		}
		filter.WithStartDateCreated(t)
	}

	if createdDate := values.Get(filterByEndCreatedDate); createdDate != "" {
		t, err := time.Parse(time.RFC3339, createdDate)
		if err != nil {
			return user.QueryFilter{}, validate.NewFieldsError(filterByEndCreatedDate, err)
		}
		filter.WithEndCreatedDate(t)
	}

	if name := values.Get(filterByName); name != "" {
		filter.WithName(name)
	}

	/* there are some validation tags in the QueryFilter model in business layer! So we call the validate() method here,
	before sending this to business layer. */
	if err := filter.Validate(); err != nil {
		return user.QueryFilter{}, err
	}

	return filter, nil
}

// ==============================================================================

func parseSummaryFilter(r *http.Request) (summary.QueryFilter, error) {
	values := r.URL.Query()

	var filter summary.QueryFilter

	if userID := values.Get("user_id"); userID != "" {
		id, err := uuid.Parse(userID)
		if err != nil {
			return summary.QueryFilter{}, validate.NewFieldsError("user_id", err)
		}

		filter.WithUserID(id)
	}

	if userName := values.Get("user_name"); userName != "" {
		filter.WithUserName(userName)
	}

	return filter, nil
}
