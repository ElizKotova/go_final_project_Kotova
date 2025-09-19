package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go_final_project_Kotova/pkg/db"
)

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	writeJSON(w, map[string]string{
		"id":      strconv.FormatInt(t.ID, 10),
		"date":    t.Date,
		"title":   t.Title,
		"comment": t.Comment,
		"repeat":  t.Repeat,
	})
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	// Проверяем метод запроса
	if r.Method != http.MethodPut {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	// Универсальная структура для принятия JSON с ID как string или number
	var reqData struct {
		ID      interface{} `json:"id"` // Принимаем и строку и число
		Date    string      `json:"date"`
		Title   string      `json:"title"`
		Comment string      `json:"comment"`
		Repeat  string      `json:"repeat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	// Преобразуем ID в строку, а затем в int64
	var idStr string
	switch v := reqData.ID.(type) {
	case string:
		idStr = v
	case float64:
		idStr = fmt.Sprintf("%.0f", v)
	case int:
		idStr = fmt.Sprintf("%d", v)
	default:
		writeError(w, http.StatusBadRequest, "invalid id type")
		return
	}

	if strings.TrimSpace(idStr) == "" {
		writeError(w, http.StatusBadRequest, "missing id")
		return
	}

	// Парсим ID из строки
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	// Создаем структуру Task с правильным типом ID
	t := db.Task{
		ID:      id,
		Date:    reqData.Date,
		Title:   reqData.Title,
		Comment: reqData.Comment,
		Repeat:  reqData.Repeat,
	}

	// Валидируем и обрабатываем задачу
	if err := validateAndProcessTask(&t, true); err != nil {
		// Определяем код ответа в зависимости от типа ошибки
		code := http.StatusBadRequest
		if err.Error() == "internal error: failed to parse today's date" {
			code = http.StatusInternalServerError
		}
		writeError(w, code, err.Error())
		return
	}

	if err := db.UpdateTask(&t); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	fmt.Printf("Successfully updated task ID %d: %+v\n", t.ID, t)
	writeJSON(w, map[string]any{})
}

// todayTimeMust парсит дату в формате dateLayout, игнорируя ошибки
func todayTimeMust(s string) time.Time {
	t, _ := time.Parse(dateLayout, s)
	return t
}
