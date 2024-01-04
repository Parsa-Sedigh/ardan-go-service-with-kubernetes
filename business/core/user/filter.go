package user

import (
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/validate"
	"github.com/google/uuid"
	"net/mail"
	"time"
)

// QueryFilter holds the available fields a query can be filtered on.
// We are using pointer semantics because the With API mutates the value(look at the with... funcs which mutate the pointer receiver that they
// have).
/* Yes the fields here are exported so you can set them directly without using with... funcs(look the methods in this file).

Note: We didn't put validation for most of the fields, because we've already validated them by using the type system(like having mail.Address).*/
type QueryFilter struct {
	ID               *uuid.UUID
	Name             *string `validate:"omitempty,min=3"`
	Email            *mail.Address
	StartCreatedDate *time.Time
	EndCreatedDate   *time.Time
}

// Validate can perform a check of the data against the validate tags.
func (qf *QueryFilter) Validate() error {
	if err := validate.Check(qf); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	return nil
}

// WithUserID sets the ID field of the QueryFilter value.
func (qf *QueryFilter) WithUserID(userID uuid.UUID) {
	qf.ID = &userID
}

// WithName sets the Name field of the QueryFilter value.
func (qf *QueryFilter) WithName(name string) {
	qf.Name = &name
}

// WithEmail sets the Email field of the QueryFilter value.
func (qf *QueryFilter) WithEmail(email mail.Address) {
	qf.Email = &email
}

// WithStartDateCreated sets the DateCreated field of the QueryFilter value.
func (qf *QueryFilter) WithStartDateCreated(startDate time.Time) {
	d := startDate.UTC()
	qf.StartCreatedDate = &d
}

// WithEndCreatedDate sets the DateCreated field of the QueryFilter value.
func (qf *QueryFilter) WithEndCreatedDate(endDate time.Time) {
	d := endDate.UTC()
	qf.EndCreatedDate = &d
}
