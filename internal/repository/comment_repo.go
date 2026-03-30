package repository

import (
	"blog-api/internal/model"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

// CommentRepo представляет репозиторий для работы с комментариями
type CommentRepo struct {
	db     *sql.DB
	logger *log.Logger
}

// NewCommentRepo создает новый репозиторий комментариев
func NewCommentRepo(db *sql.DB, logger *log.Logger) *CommentRepo {
	return &CommentRepo{db: db, logger: logger}
}

// Create создает новый комментарий
func (r *CommentRepo) Create(ctx context.Context, comment *model.Comment) error {

	query := `
		INSERT INTO comments (content, post_id, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now
	row := r.db.QueryRowContext(ctx, query,
		comment.Content,
		comment.PostID,
		comment.AuthorID,
		comment.CreatedAt,
		comment.UpdatedAt)
	err := row.Scan(&comment.ID)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("failed to write comment data to db, authorID = %d/n", comment.AuthorID)
		return err
	case err != nil:
		r.logger.Printf("Failed create comment %d: %v/n", comment.AuthorID, err)
		return err
	}
	return nil
}

// GetByID получает комментарий по ID
func (r *CommentRepo) GetByID(ctx context.Context, id int) (*model.Comment, error) {

	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comments
		WHERE id = $1
	`
	var comment model.Comment
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&comment.ID,
		&comment.Content,
		&comment.PostID,
		&comment.AuthorID,
		&comment.CreatedAt,
		&comment.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("No such comment, id = %d/n", id)
		return &model.Comment{}, ErrCommentNotFound
	case err != nil:
		r.logger.Printf("Error fetching comment with id = %d: %v/n", id, err)
		return &model.Comment{}, err
	}
	return &comment, nil
}

// GetByPostID получает комментарии к посту с пагинацией
func (r *CommentRepo) GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error) {

	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		r.logger.Printf("Error fetching comments, postID = %d, error message: %v/n", postID, err)
		return nil, ErrCommentNotFound
	}
	defer rows.Close()

	// TODO: Итерировать по результатам
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.Content, &comment.PostID, &comment.AuthorID, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			r.logger.Printf("Error fetching comments, error message: %v/n", err)
		} else {
			comments = append(comments, &comment)
		}
	}
	if comments != nil {
		return comments, nil
	}
	return nil, fmt.Errorf("Fetched 0 comments")
}

// GetCountByPostID получает количество комментариев к посту
func (r *CommentRepo) GetCountByPostID(ctx context.Context, postID int) (int, error) {

	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		r.logger.Printf("Error fetching commens count: %v/n", err)
		return 0, err
	}
	return count, nil
}

// GetByPostID получает комментарии автора с пагинацией
func (r *CommentRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Comment, error) {
	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comments
		WHERE author_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		r.logger.Printf("Error fetching comments, authorID = %d, error message: %v/n", authorID, err)
		return nil, ErrCommentNotFound
	}
	defer rows.Close()

	// TODO: Итерировать по результатам
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(&comment.ID, &comment.Content, &comment.PostID, &comment.AuthorID, &comment.CreatedAt, &comment.UpdatedAt)
		if err != nil {
			r.logger.Printf("Error fetching comments, error message: %v/n", err)
		} else {
			comments = append(comments, &comment)
		}
	}
	if comments != nil {
		return comments, nil
	}
	return nil, fmt.Errorf("Fetched 0 comments")
}

// GetCountByPostID получает количество комментариев к посту
func (r *CommentRepo) GetCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	query := `SELECT COUNT(*) FROM comments WHERE author_id = $1`

	var count int
	err := r.db.QueryRowContext(ctx, query, authorID).Scan(&count)
	if err != nil {
		r.logger.Printf("Error fetching commens count: %v/n", err)
		return 0, err
	}
	return count, nil
}

// Update обновляет комментарий
func (r *CommentRepo) Update(ctx context.Context, comment *model.Comment) error {

	query := `
		UPDATE comments
		SET content = $1, updated_at = $2
		WHERE id = $3
	`

	comment.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		comment.Content,
		comment.UpdatedAt,
		comment.ID)
	if err != nil {
		r.logger.Printf("Error updating comment %d: %v", comment.ID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error updating comment %d: %v", comment.ID, err)
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("expected to affect 1 row, affected %d", rowsAffected)
	}
	return nil
}

// Delete удаляет комментарий
func (r *CommentRepo) Delete(ctx context.Context, id int) error {

	query := `DELETE FROM comments WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Printf("Error deleting comment %d: %v", id, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error deleting comment %d: %v", id, err)
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Comment with id = %d wasn't deleted (not exist?)", id)
	}
	return nil
}

// Exists проверяет существование поста
func (r *CommentRepo) Exists(ctx context.Context, id int) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM comments WHERE id = $1)`

	row := r.db.QueryRowContext(ctx, query, id)
	var exist bool
	err := row.Scan(&exist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		} else {
			return exist, err
		}
	}
	return exist, err
}
