package repository

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard-api/internal/domain"
	"taskboard-api/internal/errs"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project domain.Project) (domain.Project, error) {
	query, args, err := r.psql.
		Insert("projects").
		Columns("name", "description").
		Values(project.Name, project.Description).
		Suffix("RETURNING id, name, description, status, created_at").
		ToSql()
	if err != nil {
		return domain.Project{}, fmt.Errorf("build create project query: %w", err)
	}

	var created domain.Project

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.Name,
		&created.Description,
		&created.Status,
		&created.CreatedAt,
	)
	if err != nil {
		return domain.Project{}, fmt.Errorf("create project: %w", err)
	}

	return created, nil
}

func (r *ProjectRepository) GetByID(ctx context.Context, id domain.ProjectID) (domain.Project, error) {
	query, args, err := r.psql.
		Select("id", "name", "description", "status", "created_at").
		From("projects").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.Project{}, fmt.Errorf("build get project query: %w", err)
	}

	var project domain.Project

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&project.ID,
		&project.Name,
		&project.Description,
		&project.Status,
		&project.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Project{}, fmt.Errorf("get project by id: %w", errs.ErrNotFound)
	}
	if err != nil {
		return domain.Project{}, fmt.Errorf("get project by id: %w", err)
	}

	return project, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]domain.Project, error) {
	query, args, err := r.psql.
		Select("id", "name", "description", "status", "created_at").
		From("projects").
		OrderBy("id DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list projects query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}
	defer rows.Close()

	var projects []domain.Project

	for rows.Next() {
		var project domain.Project

		if err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.Description,
			&project.Status,
			&project.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) Close(ctx context.Context, id domain.ProjectID) error {
	query, args, err := r.psql.
		Update("projects").
		Set("status", domain.ProjectStatusClosed).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build close project query: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("close project: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("close project: %w", errs.ErrNotFound)
	}

	return nil
}

func (r *ProjectRepository) AddMember(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
	role domain.ProjectRole,
) error {
	query, args, err := r.psql.
		Insert("project_members").
		Columns("project_id", "user_id", "role").
		Values(projectID, userID, role).
		Suffix("ON CONFLICT (project_id, user_id) DO NOTHING").
		ToSql()
	if err != nil {
		return fmt.Errorf("build add project member query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("add project member: %w", err)
	}

	return nil
}

func (r *ProjectRepository) IsMember(
	ctx context.Context,
	projectID domain.ProjectID,
	userID domain.UserID,
) (bool, error) {
	query, args, err := r.psql.
		Select("1").
		From("project_members").
		Where(sq.Eq{
			"project_id": projectID,
			"user_id":    userID,
		}).
		Limit(1).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("build is member query: %w", err)
	}

	var exists int

	err = r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check project member: %w", err)
	}

	return true, nil
}
