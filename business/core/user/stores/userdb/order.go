package userdb

import (
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/core/user"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/data/order"
)

var orderByFields = map[string]string{
	/* Note: We don't want the DB column name("user_id) to be in the domain layer. We connect the business layer with
	storage layer, here. We don't want to mix these layers together. We only connect them here, not in the business layer.*/
	user.OrderByID:      "user_id",
	user.OrderByName:    "name",
	user.OrderByEmail:   "email",
	user.OrderByRoles:   "roles",
	user.OrderByEnabled: "enabled",
}

func orderByClause(orderBy order.By) (string, error) {
	by, exists := orderByFields[orderBy.Field]
	if !exists {
		return "", fmt.Errorf("field %q does not exist", orderBy.Field)
	}

	return " ORDER BY " + by + " " + orderBy.Direction, nil
}
