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
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// UserRepo представляет репозиторий для работы с пользователями
type UserRepo struct {
	db     *sql.DB
	logger *log.Logger
}

// NewUserRepo создает новый репозиторий пользователей
func NewUserRepo(db *sql.DB, logger *log.Logger) *UserRepo {

	return &UserRepo{db: db, logger: logger}
}

// Create создает нового пользователя
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {

	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	row := r.db.QueryRowContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt)
	err := row.Scan(&user.ID)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("Failed to write user data to db, user = %s/n", user.Username)
		return err
	case err != nil:
		r.logger.Printf("Failed create user %s: %v/n", user.Username, err)
		return err
	}
	return nil
}

// GetByID получает пользователя по ID
func (r *UserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("No such user, id = %d/n", id)
		return &model.User{}, ErrUserNotFound
	case err != nil:
		r.logger.Printf("Error fetching user  with id = %d: %v/n", id, err)
		return &model.User{}, err
	}

	return &user, nil
}

// GetByEmail получает пользователя по email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {

	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var user model.User
	row := r.db.QueryRowContext(ctx, query, email)
	err := row.Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("No such user, email = %s/n", email)
		return &model.User{}, ErrUserNotFound
	case err != nil:
		r.logger.Printf("Error fetching user  with email = %s: %v/n", email, err)
		return &model.User{}, err
	}
	return &user, nil
}

// GetByUsername получает пользователя по username
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {

	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE username = $1
	`

	var user model.User
	row := r.db.QueryRowContext(ctx, query, username)
	err := row.Scan(&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt)
	switch {
	case err == sql.ErrNoRows:
		r.logger.Printf("No such user, username = %s/n", username)
		return &model.User{}, ErrUserNotFound
	case err != nil:
		r.logger.Printf("Error fetching user  with username = %s: %v/n", username, err)
		return &model.User{}, err
	}
	return &user, nil
}

// ExistsByEmail проверяет существование пользователя по email
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	row := r.db.QueryRowContext(ctx, query, email)
	var exist bool
	err := row.Scan(&exist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// fmt.Printf("%v", err)
			return false, nil
		} else {
			return exist, err
		}
	}
	return exist, err

}

// ExistsByUsername проверяет существование пользователя по username
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	row := r.db.QueryRowContext(ctx, query, username)
	var exist bool
	err := row.Scan(&exist)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// fmt.Printf("%v", err)
			return false, nil
		} else {
			return exist, err
		}
	}
	return exist, err
}

// Update обновляет данные пользователя
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {

	query := `
		UPDATE users SET 
		username = $1,
		email = $2,
		password = $3,
		created_at = $4,
		updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
		user.ID)
	if err != nil {
		r.logger.Printf("Error updating data for user id %d: %v", user.ID, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error updating user data, id %d: %v", user.ID, err)
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("User with id = %d wasn't updated (not exist?)", user.ID)
	}
	return nil
}

// Delete удаляет пользователя
func (r *UserRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Printf("Error deleting user with id %d: %v", id, err)
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Printf("Error deleting user with id %d: %v", id, err)
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("User with id = %d wasn't deleted (not exist?)", id)
	}
	return nil
}
