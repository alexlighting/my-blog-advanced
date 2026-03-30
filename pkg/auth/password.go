package auth

import (
	"crypto/rand"
	"errors"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrEmptyPassword    = errors.New("password cannot be empty")
	ErrPasswordTooShort = errors.New("password is too short")
)

const PASSWORD_MIN_LENGTH = 6
const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()-=_+{}"

// HashPassword хеширует пароль используя bcrypt
func HashPassword(password string) (string, error) {

	if password == "" {
		return "", ErrEmptyPassword
	}
	if len(password) < PASSWORD_MIN_LENGTH {
		return "", ErrPasswordTooShort
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPassword проверяет соответствие пароля и его хеша
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// GenerateRandomPassword генерирует случайный пароль (опциональное задание)
func GenerateRandomPassword(length int) (string, error) {

	if length < 6 {
		return "", errors.New("can't generate such short password")
	}
	b := make([]byte, length)
	for i := range b {
		// Int cannot return an error when using rand.Reader.
		a, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letterBytes))))
		b[i] = letterBytes[a.Int64()]
	}
	return string(b), nil
}
