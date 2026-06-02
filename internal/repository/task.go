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

type TaskRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *TaskRepository) Create(ctx context.Context, task domain.Task) (domain.Task, error) {
	query, args, err := r.psql.
		Insert("tasks").
		Columns("project_id", "assignee_id", "title", "description").
		Values(task.ProjectID, task.AssigneeID, task.Title, task.Description).
		Suffix("RETURNING id, project_id, assignee_id, title, description, status, created_at, updated_at").
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("build create task query: %w", err)
	}

	var created domain.Task

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.ProjectID,
		&created.AssigneeID,
		&created.Title,
		&created.Description,
		&created.Status,
		&created.CreatedAt,
		&created.UpdatedAt,
	)
	if err != nil {
		return domain.Task{}, fmt.Errorf("create task: %w", err)
	}

	return created, nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id domain.TaskID) (domain.Task, error) {
	query, args, err := r.psql.
		Select("id", "project_id", "assignee_id", "title", "description", "status", "created_at", "updated_at").
		From("tasks").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return domain.Task{}, fmt.Errorf("build get task query: %w", err)
	}

	var task domain.Task

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&task.ID,
		&task.ProjectID,
		&task.AssigneeID,
		&task.Title,
		&task.Description,
		&task.Status,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Task{}, fmt.Errorf("get task by id: %w", errs.ErrNotFound)
	}
	if err != nil {
		return domain.Task{}, fmt.Errorf("get task by id: %w", err)
	}

	return task, nil
}

func (r *TaskRepository) List(ctx context.Context) ([]domain.Task, error) {
	query, args, err := r.psql.
		Select("id", "project_id", "assignee_id", "title", "description", "status", "created_at", "updated_at").
		From("tasks").
		OrderBy("id DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list tasks query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.Task

	for rows.Next() {
		var task domain.Task

		if err := rows.Scan(
			&task.ID,
			&task.ProjectID,
			&task.AssigneeID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.CreatedAt,
			&task.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tasks: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id domain.TaskID, status domain.TaskStatus) error {
	query, args, err := r.psql.
		Update("tasks").
		Set("status", status).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update task status query: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update task status: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("update task status: %w", errs.ErrNotFound)
	}

	return nil
}

func (r *TaskRepository) Delete(ctx context.Context, id domain.TaskID) error {
	query, args, err := r.psql.
		Delete("tasks").
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build delete task query: %w", err)
	}

	tag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("delete task: %w", errs.ErrNotFound)
	}

	return nil
}
