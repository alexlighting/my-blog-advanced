package model

import (
	"context"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
)

// User представляет модель пользователя в системе
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"-" db:"password"` // Хешированный пароль, не отдаем в JSON
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Post представляет модель поста в блоге
type Post struct {
	ID         int       `json:"id" db:"id" validate:"numeric" redis:"id"`
	Title      string    `json:"title" db:"title" validate:"required,min=1,max=200" redis:"title"`
	Content    string    `json:"content" db:"content" validate:"required,min=1,max=100000" redis:"content"`
	AuthorID   int       `json:"author_id" db:"author_id" validate:"numeric" redis:"author_id"`
	Draft      bool      `json:"draft" db:"draft" validate:"boolean" redis:"draft"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" validate:"omitempty" redis:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" validate:"omitempty" redis:"updated_at"`
	Publish_at time.Time `json:"publish_at" db:"publish_at" validate:"omitempty" redis:"publish_at"`
}

// Comment представляет модель комментария к посту
type Comment struct {
	ID        int       `json:"id" db:"id" validate:"numeric"`
	Content   string    `json:"content" db:"content" validate:"required,min=1,max=1000"`
	PostID    int       `json:"post_id" db:"post_id" validate:"numeric"`
	AuthorID  int       `json:"author_id" db:"author_id" validate:"numeric"`
	CreatedAt time.Time `json:"created_at" db:"created_at" validate:"omitempty"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" validate:"omitempty"`
}

// UserCreateRequest представляет запрос на создание пользователя
type UserCreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=255,strongpassword"`
}

// UserLoginRequest представляет запрос на вход пользователя
type UserLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=255,strongpassword"`
}

// PostCreateRequest представляет запрос на создание поста
type PostCreateRequest struct {
	Title      string    `json:"title" validate:"required,min=1,max=200"`
	Content    string    `json:"content" validate:"required,min=1,max=100000"`
	Publish_at time.Time `json:"publish_at" validate:"omitempty"`
}

// PostUpdateRequest представляет запрос на обновление поста
type PostUpdateRequest struct {
	Title      string    `json:"title" validate:"required,min=1,max=200"`
	Content    string    `json:"content" validate:"required,min=1,max=100000"`
	Publish_at time.Time `json:"publish_at" validate:"omitempty"`
}

// CommentCreateRequest представляет запрос на создание комментария
type CommentCreateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// CommentCreateRequest представляет запрос на создание комментария
type CommentUpdateRequest struct {
	Content string `json:"content" validate:"required,min=1,max=1000"`
}

// TODO: Добавить следующие структуры и методы:

// UserResponse - структура для ответа с данными пользователя (без пароля)
type UserResponse struct {
	ID        int
	Username  string
	Email     string
	CreatedAt time.Time
}

// TokenResponse - структура для ответа с JWT токеном
type TokenResponse struct {
	Token     string
	ExpiresAt time.Time
	User      UserResponse
}

// PostResponse - структура для ответа с данными поста
type PostResponse struct {
	ID         int
	Title      string
	Content    string
	Author     UserResponse
	Draft      bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Publish_at time.Time
}

// CommentResponse - структура для ответа с данными комментария
type CommentResponse struct {
	ID        int
	Content   string
	PostID    int
	Author    UserResponse
	CreatedAt time.Time
	UpdatedAt time.Time
}

// User.ToResponse() UserResponse - преобразует User в UserResponse
func (user User) ToResponse() *UserResponse {
	return &UserResponse{ID: user.ID, Username: user.Username, Email: user.Email, CreatedAt: user.CreatedAt}
}

// Post.CanBeEditedBy(userID int) bool - проверяет, может ли пользователь редактировать пост
func (post Post) CanBeEditedBy(userID int) bool {
	return post.AuthorID == userID
}

// Post.CanBeDeletedBy(userID int) bool - проверяет, может ли пользователь удалить пост
func (post Post) CanBeDeletedBy(userID int) bool {
	return post.AuthorID == userID
}

// Comment.CanBeEditedBy(userID int) bool - проверяет, может ли пользователь редактировать комментарий
func (comment Comment) CanBeEditedBy(userID int) bool {
	return comment.AuthorID == userID
}

// Comment.CanBeDeletedBy(userID int) bool - проверяет, может ли пользователь удалить комментарий
func (comment Comment) CanBeDeletedBy(userID int) bool {
	return comment.AuthorID == userID
}

func (p *Post) ToRedisSet(ctx context.Context, db *redis.Client, key string) error {
	// Получаем элементы структуры
	val := reflect.ValueOf(p).Elem()

	// Создаем функцию для записи структуры в хранилище
	settter := func(pipe redis.Pipeliner) error {
		// Итерируемся по полям структуры
		for i := 0; i < val.NumField(); i++ {
			field := val.Type().Field(i)
			// Получаем содержимое тэга redis
			tag := field.Tag.Get("redis")
			// Записываем значение поля и содержимое тэга redis в хранилище
			if err := pipe.HSet(ctx, key, tag, val.Field(i).Interface()).Err(); err != nil {
				return err
			}
		}
		// Задаем время хранения 2 минуты
		if err := pipe.Expire(ctx, key, 2*time.Minute).Err(); err != nil {
			return err
		}
		return nil
	}

	// Сохраняем структуру в хранилище
	if _, err := db.Pipelined(ctx, settter); err != nil {
		return err
	}

	return nil
}
