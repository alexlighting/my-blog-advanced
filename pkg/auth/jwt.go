package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims представляет данные, хранимые в JWT токене
type Claims struct {
	UserID   int    `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	jwt.RegisteredClaims
	// TODO: Добавить стандартные JWT claims
	// Подсказка: используйте jwt.RegisteredClaims или jwt.StandardClaims
}

// JWTManager управляет созданием и валидацией JWT токенов
type JWTManager struct {
	secretKey []byte
	ttl       time.Duration
}

// NewJWTManager создает новый экземпляр JWT менеджера
func NewJWTManager(secretKey string, ttlHours int) *JWTManager {

	jwtManager := JWTManager{secretKey: []byte(secretKey), ttl: time.Duration(ttlHours) * time.Hour}
	return &jwtManager
}

// GenerateToken создает новый JWT токен для пользователя
func (m *JWTManager) GenerateToken(userID int, email, username string) (string, time.Time, error) {

	if username == "" || email == "" {
		return "", time.Time{}, errors.New("Can't generate token - username and email can't be empty")
	}

	tokenID := uuid.New().String()
	claims := Claims{UserID: userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			Subject:   username,
			Issuer:    "blog-api",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.ttl)),
			Audience:  []string{"my-frontend"}},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, time.Now().Add(m.ttl), nil
}

// ValidateToken проверяет и парсит JWT токен
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {

	claims := &Claims{}
	//парсим токен
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return m.secretKey, nil
		})
	//если возникли ошибки при парсинге выходим с сообщением об ошибке
	if err != nil {
		return &Claims{}, ErrInvalidToken
	}
	//если токен распарсился без ошибок извлекаем из него claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return &Claims{}, ErrInvalidToken
	} else {
		//если удалось извлечь, проверяем не истек ли срок жизни токена
		if claims.ExpiresAt.Before(time.Now()) {
			return &Claims{}, ErrExpiredToken
		}
		return claims, nil
	}

}

// RefreshToken обновляет существующий токен
func (m *JWTManager) RefreshToken(tokenString string) (string, time.Time, error) {
	claims, error := m.ValidateToken(tokenString)
	if error != nil {
		return "", time.Time{}, errors.New("failed to refresh token")
	}
	return m.GenerateToken(claims.UserID, claims.Email, claims.Username)
}
