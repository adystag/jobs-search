package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/adystag/jobs-search/internal"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func (r userRepository) GetUserByUsername(ctx context.Context, username string) (internal.User, error) {
	var user internal.User

	query := `
		SELECT
			id,
			username,
			password,
			created_at,
			updated_at
		FROM users
		WHERE username = ?
		LIMIT 1
	`
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = fmt.Errorf("%w: %w", internal.ErrUserNotFound, err)
		}

		return internal.User{}, fmt.Errorf("querying mysql users table: %w", err)
	}

	return user, nil
}

func (r userRepository) StoreUser(ctx context.Context, user *internal.User) (err error) {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("initializing mysql db transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	{
		query := `
			INSERT INTO users (username, password, created_at, updated_at)
			VALUES (?, ?, ?, ?)
		`
		args := []interface{}{
			user.Username,
			user.Password,
			user.CreatedAt,
			user.UpdatedAt,
		}

		if user.ID > 0 {
			query = `
				UPDATE users
				SET
					username = ?,
					password = ?,
					updated_at = ?
				WHERE id = ?
			`
			args = []interface{}{
				user.Username,
				user.Password,
				user.UpdatedAt,
				user.ID,
			}
		}

		stmt, err := tx.PrepareContext(ctx, query)
		if err != nil {
			return fmt.Errorf("preparing mysql query: %w", err)
		}

		defer stmt.Close()

		res, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			return fmt.Errorf("executing mysql query: %w", err)
		}

		if user.ID <= 0 {
			lastInsertedID, err := res.LastInsertId()
			if err != nil {
				return fmt.Errorf("getting last inserted id: %w", err)
			}

			user.ID = lastInsertedID
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("committing mysql db transaction: %w", err)
	}

	return nil
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}
