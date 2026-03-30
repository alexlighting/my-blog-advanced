package model

import (
	"time"
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
	ID         int       `json:"id" db:"id" validate:"numeric"`
	Title      string    `json:"title" db:"title" validate:"required,min=1,max=200"`
	Content    string    `json:"content" db:"content" validate:"required,min=1,max=100000"`
	AuthorID   int       `json:"author_id" db:"author_id" validate:"numeric"`
	Draft      bool      `json:"draft" db:"draft" validate:"boolean"`
	CreatedAt  time.Time `json:"created_at" db:"created_at" validate:"omitempty"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" validate:"omitempty"`
	Publish_at time.Time `json:"publish_at" db:"publish_at" validate:"omitempty"`
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
