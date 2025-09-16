package api

import (
	"net/http"
	"time"

	"go_final_project_Kotova/pkg/logic"
)

const dateLayout = "20060102"

// RegisterRoutes регистрирует API-эндпоинты на mux.
func RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/nextdate", func(w http.ResponseWriter, r *http.Request) {
		nowStr := r.FormValue("now")
		dateStr := r.FormValue("date")
		repeat := r.FormValue("repeat")

		var now time.Time
		var err error
		if nowStr == "" {
			now = time.Now()
		} else {
			now, err = time.Parse(dateLayout, nowStr)
			if err != nil {
				http.Error(w, "invalid now", http.StatusBadRequest)
				return
			}
		}

		next, err := logic.NextDate(now, dateStr, repeat)
		if err != nil {
			// вернуть пустую строку
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			_, _ = w.Write([]byte(""))
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte(next))
	})

	mux.HandleFunc("/api/task", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			requireAuth(addTaskHandler)(w, r)
		case http.MethodGet:
			requireAuth(getTaskHandler)(w, r)
		case http.MethodPut:
			requireAuth(editTaskHandler)(w, r)
		case http.MethodDelete:
			requireAuth(deleteTaskHandler)(w, r)
		default:
			writeJSON(w, map[string]any{"error": "method not allowed"})
		}
	})

	mux.HandleFunc("/api/task/done", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, map[string]any{"error": "method not allowed"})
			return
		}
		requireAuth(doneTaskHandler)(w, r)
	})

	mux.HandleFunc("/api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, map[string]any{"error": "method not allowed"})
			return
		}
		requireAuth(tasksHandler)(w, r)
	})

	// Authentication endpoints (no authentication required)
	mux.HandleFunc("/api/signin", signinHandler)
	mux.HandleFunc("/api/logout", logoutHandler)
}
