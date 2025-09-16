package server

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"go_final_project_Kotova/pkg/api"
	"go_final_project_Kotova/pkg/db"
)

// Start запускает HTTP-сервер.
func Start() error {
	// Инициализация БД до старта сервера
	dbFile := "scheduler.db"
	if dbPath := os.Getenv("TODO_DBFILE"); len(dbPath) > 0 {
		dbFile = dbPath
	}
	if err := db.Init(dbFile); err != nil {
		return err
	}

	// Проверяем и создаем пользователя по умолчанию при необходимости
	if err := createDefaultUser(); err != nil {
		return err
	}

	port := 7540
	if p := os.Getenv("TODO_PORT"); len(p) > 0 {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			port = v
		}
	}
	webDir := "./web"
	mux := http.NewServeMux()

	// Обслуживаем статические файлы без аутентификации
	mux.Handle("/", http.FileServer(http.Dir(webDir)))

	// Регистрируем API маршруты с аутентификацией
	api.RegisterRoutes(mux)

	addr := ":" + strconv.Itoa(port)
	log.Printf("starting static server on %s -> %s", addr, webDir)
	return http.ListenAndServe(addr, mux)
}

// createDefaultUser создает пользователя по умолчанию, если он не существует
func createDefaultUser() error {
	const query = `INSERT OR IGNORE INTO users (login, password) VALUES (?, ?)`
	_, err := db.DB.Exec(query, "user", "password")
	return err
}
