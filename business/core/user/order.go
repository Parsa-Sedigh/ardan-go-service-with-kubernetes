package user

import "github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/data/order"

// DefaultOrderBy represents the default way we sort.
var DefaultOrderBy = order.NewBy(OrderByID, order.ASC)

// Set of fields that the results can be ordered by.
// Instead of iota, we use strings because they're more readable. Integer values for these constants are not readable.
const (
	OrderByID      = "user_id"
	OrderByName    = "name"
	OrderByEmail   = "email"
	OrderByRoles   = "roles"
	OrderByEnabled = "enabled"
)
