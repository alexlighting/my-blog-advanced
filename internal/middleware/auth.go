package middleware

import (
	"blog-api/pkg/auth"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the key for storing user ID in context
	UserIDKey contextKey = "userID"
	// UserEmailKey is the key for storing user email in context
	UserEmailKey contextKey = "userEmail"
	// UserNameKey is the key for storing username in context
	UserNameKey contextKey = "username"
)

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth is a middleware that requires valid JWT token
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := extractToken(r)
		if authHeader == "" {
			writeJSONError(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		claims, err := m.jwtManager.ValidateToken(authHeader)
		if err != nil {
			writeJSONError(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
			return
		}
		//почему-то когда использовал ваши константы ничего не получалось прочитать из контекста
		// проблема решилась только после того как заменил contextKey на обычные константные строки
		ctx := context.WithValue(r.Context(), "userEmail", claims.Email)
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuth is a middleware that extracts JWT token if present, but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		authHeader := extractToken(r)
		if authHeader == "" {
			next(w, r)
			return
		}
		claims, err := m.jwtManager.ValidateToken(authHeader)
		if err != nil {
			next(w, r)
			// Log(Info, "Invalid token: %v", err)
			return
		}
		ctx := context.WithValue(r.Context(), "userEmail", claims.Email)
		ctx = context.WithValue(ctx, "userID", claims.UserID)
		ctx = context.WithValue(ctx, "username", claims.Username)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// extractToken извлекает JWT токен из заголовка Authorization
func extractToken(r *http.Request) string {
	// TODO: Извлечь JWT токен из заголовка Authorization
	// Формат: "Bearer <token>"
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return authHeader
	}
	// Проверяем формат "Bearer <token>"
	const bearerPrefix = "Bearer "
	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return ""
	}
	return strings.TrimPrefix(authHeader, bearerPrefix)
}

// GetUserIDFromContext извлекает ID пользователя из контекста
func GetUserIDFromContext(ctx context.Context) (int, bool) {
	// TODO: Извлечь userID из контекста (ключ UserIDKey)
	id, ok := ctx.Value("userID").(int)
	return id, ok
}

// GetUserEmailFromContext извлекает email пользователя из контекста
func GetUserEmailFromContext(ctx context.Context) (string, bool) {
	// TODO: Извлечь email из контекста (ключ UserEmailKey)
	email, ok := ctx.Value("userEmail").(string)
	return email, ok
}

// GetUsernameFromContext извлекает username из контекста
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	// TODO: Извлечь username из контекста (ключ UserNameKey)
	username, ok := ctx.Value("username").(string)
	return username, ok
}

// writeJSONError отправляет ошибку в формате JSON
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {
	// TODO: Отправить ошибку в формате JSON
	// Создать структуру ErrorResponse и отправить как JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	response := map[string]string{"error": message}
	json.NewEncoder(w).Encode(response)
}
