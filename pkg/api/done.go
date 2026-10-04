package api

import (
	"database/sql"
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	id := r.FormValue("id")
	if id == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJSONError(w, "задача не найдена", http.StatusNotFound)
			return
		}
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if task.Repeat == "" {
		if err := db.DeleteTask(id); err != nil {
			sendJSONError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
		return
	}

	nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := db.UpdateTaskDate(id, nextDate); err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
