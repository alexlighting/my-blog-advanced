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
	ErrPostNotFound = errors.New("post not found")
)

// PostRepo представляет репозиторий для работы с постами
type PostRepo struct {
	db     *sql.DB
	logger *log.Logger
}

// NewPostRepo создает новый репозиторий постов
func NewPostRepo(db *sql.DB, logger *log.Logger) *PostRepo {
	return &PostRepo{db: db, logger: logger}
}

// Create создает новый пост
func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	query := `
		INSERT INTO posts (title, content, author_id, draft, created_at, updated_at, publish_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now
	row := r.db.QueryRowContext(ctx, query,
		post.Title,
		post.Content,
		post.AuthorID,
		post.Draft,
		post.CreatedAt,
		post.UpdatedAt,
		post.Publish_at,
	)
	err := row.Scan(&post.ID)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("Failed to write post data to db, title = %s/n", post.Title)
		return err
	case err != nil:
		r.logger.Printf("Failed create post %s: %v/n", post.Title, err)
		return err
	}
	return nil
}

// GetByID получает пост по ID
func (r *PostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {

	query := `
		SELECT id, title, content, author_id, draft, created_at, updated_at, publish_at
		FROM posts
		WHERE id = $1
	`

	var post model.Post
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.Draft,
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Publish_at)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("No such post, id = %d/n", id)
		return &model.Post{}, ErrPostNotFound
	case err != nil:
		r.logger.Printf("Error fetching posts with id = %d: %v/n", id, err)
		return &model.Post{}, err
	}
	return &post, nil

}

// GetAll получает все посты с пагинацией
func (r *PostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	//получаем все опубликованные посты
	query := `
		SELECT id, title, content, author_id, draft, created_at, updated_at, publish_at
		FROM posts
		WHERE NOT draft
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		r.logger.Printf("Error fetching posts, limit = %d, offset = %d, error message: %v/n", limit, offset, err)
		return nil, ErrPostNotFound
	}
	defer rows.Close()

	// TODO: Итерировать по результатам
	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Draft, &post.CreatedAt, &post.UpdatedAt, &post.Publish_at)
		if err != nil {
			r.logger.Printf("Error fetching post, error message: %v/n", err)
		} else {
			posts = append(posts, &post)
		}
	}
	if posts != nil {
		return posts, nil
	}
	return nil, fmt.Errorf("Fetched 0 posts")
}

// GetTotalCount получает общее количество постов
func (r *PostRepo) GetTotalCount(ctx context.Context) (int, error) {
	//считаем все опубликованные посты
	query := `SELECT COUNT(*) FROM posts WHERE NOT draft`

	var count int
	row := r.db.QueryRowContext(ctx, query)
	err := row.Scan(&count)
	if err != nil {
		r.logger.Printf("Error fetching posts count: %v/n", err)
		return 0, err
	}
	return count, nil
}

// Update обновляет пост
func (r *PostRepo) Update(ctx context.Context, post *model.Post) error {

	query := `
		UPDATE posts
		SET title = $1, content = $2, draft = $3, updated_at = $4, publish_at = $5
		WHERE id = $6
	`
	post.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query,
		post.Title,
		post.Content,
		post.Draft,
		post.UpdatedAt,
		post.Publish_at,
		post.ID)
	if err != nil {
		r.logger.Printf("Error updating data for post %d: %v", post.ID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error updating post with id %d: %v", post.ID, err)
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Expected to affect 1 row, affected %d", rowsAffected)
	}
	return nil
}

// Delete удаляет пост
func (r *PostRepo) Delete(ctx context.Context, id int) error {

	query := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Printf("Error deleting post with id %d: %v", id, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error deleting post with id %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("Post with id = %d wasn't deleted (not exist?)", id)
	}
	return nil
}

// Exists проверяет существование поста
func (r *PostRepo) Exists(ctx context.Context, id int) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`

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

// GetByAuthorID получает посты определенного автора
func (r *PostRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	//получаем список опубликованных постов автора
	query := `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts
		WHERE author_id = $1 AND NOT draft
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	// TODO: Выполнить запрос
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		r.logger.Printf("Error fetching posts of author %d, limit = %d, offset = %d, error message: %v/n", authorID, limit, offset, err)
		return nil, ErrPostNotFound
	}
	defer rows.Close()

	// TODO: Итерировать по результатам
	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			r.logger.Printf("Error fetching post, error message: %v/n", err)
		} else {
			posts = append(posts, &post)
		}
	}
	if posts != nil {
		return posts, nil
	}
	return nil, fmt.Errorf("Fetched 0 posts")
}

// GetTotalCount получает общее количество постов
func (r *PostRepo) GetTotalCountByAuthor(ctx context.Context, authorID int) (int, error) {
	//получаем общее количество опубликованных постов автора
	query := `SELECT COUNT(*) FROM posts WHERE author_id = $1 AND NOT draft`

	var count int
	row := r.db.QueryRowContext(ctx, query, authorID)
	err := row.Scan(&count)
	if err != nil {
		r.logger.Printf("Error fetching posts count: %v/n", err)
		return 0, err
	}
	return count, nil
}

// PublishById побликует черновик в качестве статьи
func (r *PostRepo) PublishById(ctx context.Context, postID int) error {
	//получаем общее количество опубликованных постов автора
	query := `UPDATE posts
		SET draft = FALSE, publish_at = $1
		WHERE id = $2`

	// var count int
	result, err := r.db.ExecContext(ctx, query, time.Now(), postID)
	if err != nil {
		r.logger.Printf("Error delayed publishing post %d: %v", postID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error updating post with id %d: %v", postID, err)
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("Expected to affect 1 row, affected %d", rowsAffected)
	}
	return nil
}

// GetDraftToUpdate получает список всех черновиков у которых подошло время публикации
func (r *PostRepo) GetDraftToUpdate(ctx context.Context) ([]int, error) {
	//получаем все опубликованные посты
	query := `SELECT id FROM posts WHERE draft AND publish_at <= $1`
	// query := `SELECT id FROM posts WHERE draft AND publish_at <= NOW()`
	// правильнее было-бы сделать так (или даже сделать предварительной скомпилированный
	//  запрос), но я обнаружил что в go вызов внутренней функции PostgresSQL NOW()
	// работает как-то странно, в  DBEaver запрос отрабатывает нормально, а в go
	// выдается пустой rows если разница между Now и временем в базе данных невелика.
	// пришлось обходить эту проблему используя time.Now()
	rows, err := r.db.QueryContext(ctx, query, time.Now())
	if err != nil {
		r.logger.Printf("Error geting draft list: %v/n", err)
		return nil, ErrPostNotFound
	}
	defer rows.Close()

	// Итерируем по результатам и сохраняем все id в срез
	var post_id int
	var posts []int
	for rows.Next() {
		err := rows.Scan(&post_id)
		log.Printf("Rows: %v", post_id)
		if err != nil {
			r.logger.Printf("Error fetching post, error message: %v/n", err)
		} else {
			posts = append(posts, post_id)
		}
	}
	if posts != nil {
		return posts, nil
	}
	return nil, fmt.Errorf("Fetched 0 posts")
}
