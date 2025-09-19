package api

import (
	"net/http"
	"strings"
	"time"

	"go_final_project_Kotova/pkg/db"
	"go_final_project_Kotova/pkg/logic"
)

// doneTaskHandler обрабатывает POST /api/task/done
func doneTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	// Получаем задачу по ID
	task, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}

	// Если задача не периодическая (нет repeat), удаляем её
	if strings.TrimSpace(task.Repeat) == "" {
		if err := db.DeleteTask(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, map[string]any{})
		return
	}

	// Для периодической задачи рассчитываем следующую дату
	now := time.Now()
	nextDate, err := logic.NextDate(now, task.Date, task.Repeat)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Обновляем дату в базе данных
	if err := db.UpdateDate(nextDate, id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, map[string]any{})
}

// deleteTaskHandler обрабатывает DELETE /api/task
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodDelete {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	if err := db.DeleteTask(id); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, map[string]any{})
}
