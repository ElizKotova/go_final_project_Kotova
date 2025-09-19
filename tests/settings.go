package tests

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims структура для JWT токена
type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

var Port = 7540
var DBFile = "../scheduler.db"
var FullNextDate = false
var Search = false
var Token = ``

// GenerateTestToken генерирует JWT токен для целей тестирования
func GenerateTestToken() string {
	// Получаем пароль из переменной окружения
	password := os.Getenv("TODO_PASSWORD")
	if password == "" {
		// Если пароль не установлен, используем значение по умолчанию
		password = "12345"
	}

	// Создаем хэш пароля
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	// Создаем утверждения для токена
	claims := &Claims{
		PasswordHash: passwordHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Получаем секрет из переменной окружения или используем значение по умолчанию
	jwtSecret := os.Getenv("TODO_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret" // Значение по умолчанию для обратной совместимости
	}

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return ""
	}

	return tokenString
}

// EnableAuth включает аутентификацию для тестов путем установки переменной Token
func EnableAuth() {
	// Устанавливаем переменные окружения если они не установлены
	if os.Getenv("TODO_PASSWORD") == "" {
		os.Setenv("TODO_PASSWORD", "12345")
	}
	if os.Getenv("TODO_JWT_SECRET") == "" {
		os.Setenv("TODO_JWT_SECRET", "secret")
	}

	// Генерируем токен и устанавливаем его
	Token = GenerateTestToken()
}

// init автоматически включает аутентификацию при запуске тестов
func init() {
	// Всегда включаем аутентификацию для тестов
	// Устанавливаем переменные окружения по умолчанию если они не установлены
	EnableAuth()
}
