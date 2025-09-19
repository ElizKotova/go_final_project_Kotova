package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig хранит конфигурацию аутентификации
type AuthConfig struct {
	Password     string
	PasswordHash string
	JWTSecret    string
}

// GlobalAuthConfig глобальная конфигурация аутентификации
var GlobalAuthConfig *AuthConfig

// InitAuthConfig инициализирует конфигурацию аутентификации при запуске приложения
// Считывает переменные окружения TODO_PASSWORD и TODO_JWT_SECRET один раз при запуске
// и сохраняет их в глобальной переменной GlobalAuthConfig для последующего использования
func InitAuthConfig() {
	password := os.Getenv("TODO_PASSWORD")
	jwtSecret := os.Getenv("TODO_JWT_SECRET")

	// Если пароль не установлен, аутентификация отключена
	if password == "" {
		GlobalAuthConfig = nil
		return
	}

	// Если секрет не установлен, используем значение по умолчанию
	if jwtSecret == "" {
		jwtSecret = "secret"
	}

	// Создаем хэш пароля
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	GlobalAuthConfig = &AuthConfig{
		Password:     password,
		PasswordHash: passwordHash,
		JWTSecret:    jwtSecret,
	}
}

// User представляет пользователя системы
type User struct {
	ID       int64  `json:"id" db:"id"`
	Login    string `json:"login" db:"login"`
	Password string `json:"password" db:"password"`
}

// Структура Claims используется для хранения информации, которая будет включена в JWT-токен
type Claims struct {
	PasswordHash string `json:"password_hash"`
	jwt.RegisteredClaims
}

// signinHandler обрабатывает POST /api/signin
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Декодируем JSON из тела запроса
	var credentials struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Проверяем, что пароль не пустой
	credentials.Password = credentials.Password
	if credentials.Password == "" {
		writeError(w, http.StatusBadRequest, "password is required")
		return
	}

	// Проверяем длину пароля (не более 255 символов)
	if len(credentials.Password) > 255 {
		writeError(w, http.StatusBadRequest, "password is too long")
		return
	}

	// Проверяем учетные данные
	if GlobalAuthConfig == nil || credentials.Password != GlobalAuthConfig.Password {
		writeError(w, http.StatusUnauthorized, "Неверный пароль")
		return
	}

	// Генерируем JWT-токен
	tokenString, err := generateJWTToken(credentials.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	// Возвращаем токен в ответе
	writeJSON(w, map[string]any{
		"token": tokenString,
	})
}

// generateJWTToken генерирует JWT-токен с хэшем пароля
func generateJWTToken(password string) (string, error) {
	// Создаем хэш пароля
	hash := sha256.Sum256([]byte(password))
	passwordHash := hex.EncodeToString(hash[:])

	// Создаем утверждения для токена
	claims := &Claims{
		PasswordHash: passwordHash,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(8 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Создаем токен
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(GlobalAuthConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// authenticate проверяет JWT-токен
// Конфигурация аутентификации считывается на старте приложения в InitAuthConfig()
func authenticate(tokenString string) bool {
	// Проверяем, что аутентификация включена
	if GlobalAuthConfig == nil {
		return false
	}

	// Парсим токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(GlobalAuthConfig.JWTSecret), nil
	})

	// Проверяем валидность токена и соответствие хэша пароля
	if err != nil || !token.Valid {
		return false
	}

	// Проверяем, что хэш пароля в токене соответствует ожидаемому хэшу
	if claims.PasswordHash != GlobalAuthConfig.PasswordHash {
		return false
	}

	return true
}

// requireAuth оборачивает обработчик и проверяет аутентификацию
func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, включена ли аутентификация
		if GlobalAuthConfig == nil {
			// Если аутентификация не включена, пропускаем аутентификацию
			handler(w, r)
			return
		}

		// Получаем токен из cookie
		tokenCookie, err := r.Cookie("token")
		if err != nil || tokenCookie.Value == "" {
			http.Error(w, "Authentification required", http.StatusUnauthorized)
			return
		}

		// Проверяем токен
		if !authenticate(tokenCookie.Value) {
			http.Error(w, "Authentification required", http.StatusUnauthorized)
			return
		}

		// Вызываем оригинальный обработчик
		handler(w, r)
	}
}

// logoutHandler обрабатывает POST /api/logout
func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Для JWT токенов не нужно ничего удалять из БД
	// Просто возвращаем пустой ответ
	writeJSON(w, map[string]any{})
}
