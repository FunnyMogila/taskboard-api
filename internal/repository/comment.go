package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	"taskboard-api/internal/domain"
)

type CommentRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *CommentRepository) Create(ctx context.Context, comment domain.Comment) (domain.Comment, error) {
	query, args, err := r.psql.
		Insert("comments").
		Columns("task_id", "author_id", "text").
		Values(comment.TaskID, comment.AuthorID, comment.Text).
		Suffix("RETURNING id, task_id, author_id, text, created_at").
		ToSql()
	if err != nil {
		return domain.Comment{}, fmt.Errorf("build create comment query: %w", err)
	}

	var created domain.Comment

	err = r.pool.QueryRow(ctx, query, args...).Scan(
		&created.ID,
		&created.TaskID,
		&created.AuthorID,
		&created.Text,
		&created.CreatedAt,
	)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("create comment: %w", err)
	}

	return created, nil
}

func (r *CommentRepository) ListByTask(ctx context.Context, taskID domain.TaskID) ([]domain.Comment, error) {
	query, args, err := r.psql.
		Select("id", "task_id", "author_id", "text", "created_at").
		From("comments").
		Where(sq.Eq{"task_id": taskID}).
		OrderBy("id ASC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build list comments query: %w", err)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list comments by task: %w", err)
	}
	defer rows.Close()

	var comments []domain.Comment

	for rows.Next() {
		var comment domain.Comment

		if err := rows.Scan(
			&comment.ID,
			&comment.TaskID,
			&comment.AuthorID,
			&comment.Text,
			&comment.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate comments: %w", err)
	}

	return comments, nil
}
