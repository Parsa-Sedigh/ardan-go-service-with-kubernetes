// Package userdb contains user related CRUD functionality.
package userdb

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/core/user"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/data/order"
	database "github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/database/pgx"
	"github.com/Parsa-Sedigh/ardan-go-service-with-kubernetes/business/sys/database/pgx/dbarray"
	"net/mail"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Store manages the set of APIs for user database access.
type Store struct {
	log    *zap.SugaredLogger
	db     sqlx.ExtContext
	inTran bool
}

// NewStore constructs the api for data access.
func NewStore(log *zap.SugaredLogger, db *sqlx.DB) *Store {
	return &Store{
		log: log,
		db:  db,
	}
}

// WithinTran runs passed function and do commit/rollback at the end.
func (s *Store) WithinTran(ctx context.Context, fn func(s user.Storer) error) error {
	if s.inTran {
		return fn(s)
	}

	f := func(tx *sqlx.Tx) error {
		s := &Store{
			log:    s.log,
			db:     tx,
			inTran: true,
		}
		return fn(s)
	}

	return database.WithinTran(ctx, s.log, s.db.(*sqlx.DB), f)
}

// Create inserts a new user into the database.
func (s *Store) Create(ctx context.Context, usr user.User) error {
	const q = `
	INSERT INTO users
		(user_id, name, email, password_hash, roles, enabled, date_created, date_updated)
	VALUES
		(:user_id, :name, :email, :password_hash, :roles, :enabled, :date_created, :date_updated)`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return fmt.Errorf("create: %w", user.ErrUniqueEmail)
		}
		return fmt.Errorf("inserting user: %w", err)
	}

	return nil
}

// Update replaces a user document in the database.
func (s *Store) Update(ctx context.Context, usr user.User) error {
	const q = `
	UPDATE
		users
	SET 
		"name" = :name,
		"email" = :email,
		"roles" = :roles,
		"password_hash" = :password_hash,
		"date_updated" = :date_updated
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, toDBUser(usr)); err != nil {
		if errors.Is(err, database.ErrDBDuplicatedEntry) {
			return user.ErrUniqueEmail
		}
		return fmt.Errorf("updating userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Delete removes a user from the database.
func (s *Store) Delete(ctx context.Context, usr user.User) error {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: usr.ID.String(),
	}

	const q = `
	DELETE FROM
		users
	WHERE
		user_id = :user_id`

	if err := database.NamedExecContext(ctx, s.log, s.db, q, data); err != nil {
		return fmt.Errorf("deleting userID[%s]: %w", usr.ID, err)
	}

	return nil
}

// Query retrieves a list of existing users from the database.
func (s *Store) Query(ctx context.Context, filter user.QueryFilter, orderBy order.By, pageNumber int, rowsPerPage int) ([]user.User, error) {
	data := struct {
		ID          string `db:"user_id"`
		Name        string `db:"name"`
		Email       string `db:"email"`
		Offset      int    `db:"offset"`
		RowsPerPage int    `db:"rows_per_page"`
	}{
		Offset:      (pageNumber - 1) * rowsPerPage,
		RowsPerPage: rowsPerPage,
	}

	const q = `
	SELECT
		*
	FROM
		users
	`
	buf := bytes.NewBufferString(q)
	buf.WriteString(" OFFSET :offset ROWS FETCH NEXT :rows_per_page ROWS ONLY")

	var usrs []dbUser
	if err := database.NamedQuerySlice(ctx, s.log, s.db, buf.String(), data, &usrs); err != nil {
		return nil, fmt.Errorf("selecting users: %w", err)
	}

	return toCoreUserSlice(usrs)
}

// Count returns the total number of users in the DB.
func (s *Store) Count(ctx context.Context, filter user.QueryFilter) (int, error) {
	data := map[string]interface{}{}

	const q = `
	SELECT
		count(1)
	FROM
		users`

	buf := bytes.NewBufferString(q)
	applyFilter(filter, data, buf)

	var count struct {
		Count int `db:"count"`
	}
	if err := database.NamedQueryStruct(ctx, s.log, s.db, buf.String(), data, &count); err != nil {
		return 0, fmt.Errorf("namedquerystruct: %w", err)
	}

	return count.Count, nil
}

// QueryByID gets the specified user from the database.
func (s *Store) QueryByID(ctx context.Context, userID uuid.UUID) (user.User, error) {
	data := struct {
		UserID string `db:"user_id"`
	}{
		UserID: userID.String(),
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE 
		user_id = :user_id`

	var usr dbUser
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting userID[%q]: %w", userID, err)
	}

	return toCoreUser(usr)
}

// QueryByIDs gets the specified users from the database.
func (s *Store) QueryByIDs(ctx context.Context, userIDs []uuid.UUID) ([]user.User, error) {
	ids := make([]string, len(userIDs))
	for i, userID := range userIDs {
		ids[i] = userID.String()
	}

	data := struct {
		ID any `db:"user_id"`
	}{
		ID: dbarray.Array(ids),
	}

	const q = `
	SELECT
        user_id, name, email, password_hash, roles, enabled, department, date_created, date_updated
	FROM
		users
	WHERE
		user_id = ANY(:user_id)`

	var dbUsrs []dbUser
	if err := sqldb.NamedQuerySlice(ctx, s.log, s.db, q, data, &dbUsrs); err != nil {
		if errors.Is(err, sqldb.ErrDBNotFound) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("namedquerystruct: %w", err)
	}

	return toCoreUserSlice(dbUsrs)
}

// QueryByEmail gets the specified user from the database by email.
func (s *Store) QueryByEmail(ctx context.Context, email mail.Address) (user.User, error) {
	data := struct {
		Email string `db:"email"`
	}{
		Email: email.Address,
	}

	const q = `
	SELECT
		*
	FROM
		users
	WHERE
		email = :email`

	var usr dbUser
	if err := database.NamedQueryStruct(ctx, s.log, s.db, q, data, &usr); err != nil {
		if errors.Is(err, database.ErrDBNotFound) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("selecting email[%q]: %w", email, err)
	}

	return toCoreUser(usr)
}
