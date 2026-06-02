package repository

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type UserRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *UserRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	query, args, err := r.psql.
		Insert("users").
		Columns("name", "email").
		Values(user.Name, user.Email).
		Suffix("RETURNING id, name, email, created_at").
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build create user query: %w", err)
	}

	var created domain.User

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.Name,
		&created.Email,
		&created.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.User{}, fmt.Errorf("create user: %w", errs.ErrAlreadyExists)
		}

		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	return created, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id domain.UserID) (domain.User, error) {
	query, args, err := r.psql.
		Select("id", "name", "email", "created_at").
		From("users").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.User{}, fmt.Errorf("build get user query: %w", err)
	}

	var user domain.User

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, fmt.Errorf("get user by id: %w", errs.ErrNotFound)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}
