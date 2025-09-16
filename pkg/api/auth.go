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
		writeJSON(w, map[string]any{"error": "method not allowed"})
		return
	}

	// Декодируем JSON из тела запроса
	var credentials struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		writeJSON(w, map[string]any{"error": "invalid JSON"})
		return
	}

	// Проверяем, что пароль не пустой
	credentials.Password = credentials.Password
	if credentials.Password == "" {
		writeJSON(w, map[string]any{"error": "password is required"})
		return
	}

	// Проверяем длину пароля (не более 255 символов)
	if len(credentials.Password) > 255 {
		writeJSON(w, map[string]any{"error": "password is too long"})
		return
	}

	// Получаем пароль из переменной окружения
	expectedPassword := os.Getenv("TODO_PASSWORD")
	if expectedPassword == "" {
		writeJSON(w, map[string]any{"error": "authentication not configured"})
		return
	}

	// Проверяем учетные данные
	if credentials.Password != expectedPassword {
		writeJSON(w, map[string]any{"error": "Неверный пароль"})
		return
	}

	// Генерируем JWT-токен
	tokenString, err := generateJWTToken(expectedPassword)
	if err != nil {
		writeJSON(w, map[string]any{"error": "internal server error"})
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

	// Получаем секрет из переменной окружения или используем значение по умолчанию
	jwtSecret := os.Getenv("TODO_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret" // Значение по умолчанию для обратной совместимости
	}

	// Подписываем токен
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// authenticate проверяет JWT-токен
func authenticate(tokenString string) bool {
	// Получаем пароль из переменной окружения
	expectedPassword := os.Getenv("TODO_PASSWORD")
	if expectedPassword == "" {
		return false
	}

	// Создаем хэш ожидаемого пароля
	hash := sha256.Sum256([]byte(expectedPassword))
	expectedPasswordHash := hex.EncodeToString(hash[:])

	// Получаем секрет из переменной окружения или используем значение по умолчанию
	jwtSecret := os.Getenv("TODO_JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "secret" // Значение по умолчанию для обратной совместимости
	}

	// Парсим токен
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Проверяем метод подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	// Проверяем валидность токена и соответствие хэша пароля
	if err != nil || !token.Valid {
		return false
	}

	// Проверяем, что хэш пароля в токене соответствует ожидаемому хэшу
	if claims.PasswordHash != expectedPasswordHash {
		return false
	}

	return true
}

// requireAuth оборачивает обработчик и проверяет аутентификацию
func requireAuth(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, установлен ли пароль в переменной окружения
		pass := os.Getenv("TODO_PASSWORD")
		if len(pass) == 0 {
			// Если пароль не установлен, пропускаем аутентификацию
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
		writeJSON(w, map[string]any{"error": "method not allowed"})
		return
	}

	// Для JWT токенов не нужно ничего удалять из БД
	// Просто возвращаем пустой ответ
	writeJSON(w, map[string]any{})
}
