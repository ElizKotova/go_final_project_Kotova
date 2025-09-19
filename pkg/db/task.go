package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// DefaultTaskLimit это значение лимита задач по умолчанию
const DefaultTaskLimit = 50

// Task описывает задачу в планировщике
type Task struct {
	ID      int64  `json:"id" db:"id"`
	Date    string `json:"date" db:"date"`
	Title   string `json:"title" db:"title"`
	Comment string `json:"comment" db:"comment"`
	Repeat  string `json:"repeat" db:"repeat"`
}

// AddTask добавляет запись в таблицу scheduler и возвращает её идентификатор
func AddTask(task *Task) (int64, error) {
	const query = `INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// Tasks возвращает список задач, отсортированных по дате по возрастанию, ограниченных limit
func Tasks(limit int) ([]*Task, error) {
	if limit <= 0 {
		limit = DefaultTaskLimit
	}
	const query = `SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC, id ASC LIMIT ?`
	rows, err := DB.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		t := new(Task)
		if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	// Проверяем ошибки, которые могли возникнуть при итерации по строкам
	if err = rows.Err(); err != nil {
		return nil, err
	}

	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	return tasks, nil
}

// TasksWithSearch возвращает список задач с поиском, отсортированных по дате по возрастанию
func TasksWithSearch(limit int, search string) ([]*Task, error) {
	if limit <= 0 {
		limit = DefaultTaskLimit
	}

	// Если поиск не указан, возвращаем все задачи
	if strings.TrimSpace(search) == "" {
		return Tasks(limit)
	}

	// Проверяем, является ли search датой в формате 02.01.2006
	if dateTime, err := time.Parse("02.01.2006", search); err == nil {
		// Преобразуем в формат 20060102
		dateStr := dateTime.Format("20060102")
		const query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC, id ASC LIMIT ?`
		rows, err := DB.Query(query, dateStr, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var tasks []*Task
		for rows.Next() {
			t := new(Task)
			if err := rows.Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat); err != nil {
				return nil, err
			}
			tasks = append(tasks, t)
		}

		// Проверяем ошибки, которые могли возникнуть при итерации по строкам
		if err = rows.Err(); err != nil {
			return nil, err
		}

		if tasks == nil {
			tasks = make([]*Task, 0)
		}
		return tasks, nil
	}

	// Поиск по заголовку или комментарию (регистронезависимо в Go)
	// Получаем все задачи, а затем фильтруем в Go для корректной работы с Unicode
	allTasks, err := Tasks(limit * 3) // Получаем больше задач для фильтрации
	if err != nil {
		return nil, err
	}

	// Фильтруем по поисковому термину (регистронезависимо)
	searchLower := strings.ToLower(search)
	var tasks []*Task
	for _, task := range allTasks {
		titleLower := strings.ToLower(task.Title)
		commentLower := strings.ToLower(task.Comment)
		if strings.Contains(titleLower, searchLower) || strings.Contains(commentLower, searchLower) {
			tasks = append(tasks, task)
			if len(tasks) >= limit {
				break
			}
		}
	}

	if tasks == nil {
		tasks = make([]*Task, 0)
	}
	return tasks, nil
}

// GetTask возвращает задачу по идентификатору
func GetTask(id string) (*Task, error) {
	const query = `SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?`
	t := new(Task)
	err := DB.QueryRow(query, id).Scan(&t.ID, &t.Date, &t.Title, &t.Comment, &t.Repeat)
	if err != nil {
		// Если задача не найдена, возвращаем специальную ошибку
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, err
	}
	return t, nil
}

// UpdateTask обновляет поля задачи по идентификатору
func UpdateTask(task *Task) error {
	const query = `UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := DB.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("incorrect id for updating task")
	}
	return nil
}

// DeleteTask удаляет задачу по идентификатору
func DeleteTask(id string) error {
	const query = `DELETE FROM scheduler WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}

// UpdateDate обновляет только дату задачи по идентификатору
func UpdateDate(date string, id string) error {
	const query = `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := DB.Exec(query, date, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
