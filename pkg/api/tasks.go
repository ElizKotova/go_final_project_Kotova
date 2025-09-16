package api

import (
	"net/http"
	"strconv"
	"strings"

	"go_final_project_Kotova/pkg/db"
)

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")
	search = strings.TrimSpace(search)

	// Отладочная информация: проверяем, какой поисковый запрос получен
	if search != "" {
		// Можно добавить логирование для отладки
		// log.Printf("Search query: %q", search)
	}

	tasks, err := db.TasksWithSearch(50, search)
	if err != nil {
		writeJSON(w, map[string]any{"error": err.Error()})
		return
	}
	// Преобразуем в строки согласно требованиям теста
	resp := make([]map[string]string, 0, len(tasks))
	for _, t := range tasks {
		resp = append(resp, map[string]string{
			"id":      strconv.FormatInt(t.ID, 10),
			"date":    t.Date,
			"title":   t.Title,
			"comment": t.Comment,
			"repeat":  t.Repeat,
		})
	}
	writeJSON(w, map[string]any{"tasks": resp})
}
