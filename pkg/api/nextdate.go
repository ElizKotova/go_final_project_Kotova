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
	id := r.URL.Query().Get("id")
	if strings.TrimSpace(id) == "" {
		writeJSON(w, map[string]any{"error": "missing id"})
		return
	}
	t, err := db.GetTask(id)
	if err != nil {
		writeJSON(w, map[string]any{"error": "task not found"})
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
	// Универсальная структура для принятия JSON с ID как string или number
	var reqData struct {
		ID      interface{} `json:"id"` // Принимаем и строку и число
		Date    string      `json:"date"`
		Title   string      `json:"title"`
		Comment string      `json:"comment"`
		Repeat  string      `json:"repeat"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
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
		writeJSON(w, map[string]any{"error": "invalid id type"})
		return
	}

	if strings.TrimSpace(idStr) == "" {
		writeJSON(w, map[string]any{"error": "missing id"})
		return
	}

	// Парсим ID из строки
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, map[string]any{"error": "invalid id"})
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
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}

	if err := db.UpdateTask(&t); err != nil {
		writeJSON(w, map[string]any{"error": fmt.Sprint(err)})
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
