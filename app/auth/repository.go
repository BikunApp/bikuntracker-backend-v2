package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) GetUser(npm string) (res models.User, err error) {
	ctx := context.Background()

	row := r.db.QueryRow(
		ctx,
		`SELECT * FROM account WHERE npm=$1;`,
		npm,
	)
	err = row.Scan(&res.Id, &res.Name, &res.Npm, &res.Email, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		err = fmt.Errorf("unable to execute get user SQL: %w", err)
		return
	}

	return
}

func (r *repository) GetOrCreateUser(name, npm, email string) (res models.User, err error) {
	ctx := context.Background()

	user, err := r.GetUser(npm)
	if err == nil {
		return user, nil
	}

	if !errors.Is(err, pgx.ErrNoRows) {
		err = fmt.Errorf("unable to execute get user SQL: %w", err)
		return
	}

	row := r.db.QueryRow(
		ctx,
		`INSERT INTO account (name, npm, email)
    VALUES ($1, $2, $3)
    RETURNING *;
  `,
		name,
		npm,
		email,
	)
	err = row.Scan(&res.Id, &res.Name, &res.Npm, &res.Email, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		err = fmt.Errorf("unable to execute create user SQL: %w", err)
		return
	}

	return
}
