package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go_final_project_Kotova/pkg/db"
	"go_final_project_Kotova/pkg/logic"
)

// afterNow возвращает true, если a > b
func afterNow(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	if ay != by {
		return ay > by
	}
	if am != bm {
		return am > bm
	}
	return ad > bd
}

// validateAndProcessTask проверяет и обрабатывает поля задачи
func validateAndProcessTask(t *db.Task, isEdit bool) error {
	t.Title = strings.TrimSpace(t.Title)
	if t.Title == "" {
		return fmt.Errorf("title is required")
	}

	now := time.Now()
	todayStr := now.Format(dateLayout)
	// если дата пустая — берём сегодня
	if strings.TrimSpace(t.Date) == "" {
		t.Date = todayStr
	}
	// сначала проверяем формат даты
	_, err := time.Parse(dateLayout, t.Date)
	if err != nil {
		return fmt.Errorf("bad date")
	}
	// если указано правило — валидируем через NextDate (оно же вернёт ошибку при неподдерживаемых/неверных форматах)
	if len(strings.TrimSpace(t.Repeat)) > 0 {
		// фиксируем now по полуночи текущего дня, чтобы избежать пограничных эффектов
		todayTime, err := time.Parse(dateLayout, todayStr)
		if err != nil {
			return fmt.Errorf("internal error: failed to parse today's date")
		}
		if _, err = logic.NextDate(todayTime, t.Date, t.Repeat); err != nil {
			return err
		}
	}

	// Для новых задач и редактирования задач
	if isEdit || t.Date < todayStr {
		if t.Repeat == "" {
			t.Date = todayStr
		} else {
			// Для периодических задач вычисляем следующую дату
			todayTime, err := time.Parse(dateLayout, todayStr)
			if err != nil {
				return fmt.Errorf("internal error: failed to parse today's date")
			}
			if isEdit {
				// Для редактирования используем специальную функцию
				next, err := logic.NextDate(todayTimeMust(todayStr), t.Date, t.Repeat)
				if err != nil {
					return err
				}
				t.Date = next
			} else {
				// Для новых задач
				next, err := logic.NextDate(todayTime, t.Date, t.Repeat)
				if err != nil {
					return err
				}
				t.Date = next
			}
		}
	}

	return nil
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	var t db.Task
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	if err := validateAndProcessTask(&t, false); err != nil {
		// Определяем код ответа в зависимости от типа ошибки
		code := http.StatusBadRequest
		if err.Error() == "internal error: failed to parse today's date" {
			code = http.StatusInternalServerError
		}
		writeError(w, code, err.Error())
		return
	}

	id, err := db.AddTask(&t)
	if err != nil {
		// Определяем код ответа в зависимости от типа ошибки
		code := http.StatusInternalServerError
		// Можно добавить дополнительную логику для определения кода ответа
		// в зависимости от типа ошибки базы данных
		writeError(w, code, err.Error())
		return
	}
	writeJSON(w, map[string]any{"id": id})
}
