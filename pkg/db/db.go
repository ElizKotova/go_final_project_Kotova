package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func Init(dbFile string) error {
	database, err := sql.Open("sqlite", dbFile)
	if err != nil {
		return err
	}
	if err = database.Ping(); err != nil {
		_ = database.Close()
		return err
	}
	DB = database

	// Настройка базы данных для регистронезависимого поиска
	_, err = DB.Exec("PRAGMA case_sensitive_like = OFF")
	if err != nil {
		log.Printf("Warning: Failed to set case_sensitive_like = OFF: %v", err)
	}

	// Всегда проверяем, что таблицы существуют
	// Создаем таблицу планировщика
	schedulerSchema := `CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
		title VARCHAR(255) NOT NULL DEFAULT "",
		comment TEXT NOT NULL DEFAULT "",
		repeat VARCHAR(128) NOT NULL DEFAULT ""
	)`

	if _, err = DB.Exec(schedulerSchema); err != nil {
		log.Printf("Error creating scheduler table: %v", err)
		return fmt.Errorf("failed to create scheduler table: %w", err)
	}

	// Создаем индекс для таблицы планировщика
	indexSchema := `CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date)`
	if _, err = DB.Exec(indexSchema); err != nil {
		log.Printf("Error creating index: %v", err)
		return fmt.Errorf("failed to create index: %w", err)
	}

	// Создаем таблицу пользователей
	usersSchema := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		login VARCHAR(255) NOT NULL UNIQUE,
		password VARCHAR(255) NOT NULL,
		token VARCHAR(255) NOT NULL DEFAULT ""
	)`

	if _, err = DB.Exec(usersSchema); err != nil {
		log.Printf("Error creating users table: %v", err)
		return fmt.Errorf("failed to create users table: %w", err)
	}

	log.Println("Database tables created successfully")

	// Создаем пользователя по умолчанию, если он не существует
	const query = `INSERT OR IGNORE INTO users (login, password) VALUES (?, ?)`
	_, err = DB.Exec(query, "user", "password")
	if err != nil {
		log.Printf("Warning: Failed to create default user: %v", err)
	}

	return nil
}

func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
