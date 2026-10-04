package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go_final_project/pkg/db"
)

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		addTask(w, r)
	case http.MethodGet:
		getTask(w, r)
	case http.MethodPut:
		updateTask(w, r)
	case http.MethodDelete:
		deleteTask(w, r)
	default:
		sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
	}
}

func addTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendJSONError(w, "не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" {
		task.Date = today
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		sendJSONError(w, "неверный формат даты", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendJSONError(w, "неверный формат правила повторения", http.StatusBadRequest)
			return
		}
		if parsedDate.Format("20060102") < today {
			task.Date = next
		}
	} else if parsedDate.Format("20060102") < today {
		task.Date = today
	}

	id, err := db.AddTask(&task)
	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"id": strconv.FormatInt(id, 10)})
}

func getTask(w http.ResponseWriter, r *http.Request) {
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

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

func updateTask(w http.ResponseWriter, r *http.Request) {
	var task db.Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
		return
	}

	if task.ID == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}

	if task.Title == "" {
		sendJSONError(w, "не указан заголовок задачи", http.StatusBadRequest)
		return
	}

	now := time.Now()
	today := now.Format("20060102")

	if task.Date == "" {
		task.Date = today
	}

	parsedDate, err := time.Parse("20060102", task.Date)
	if err != nil {
		sendJSONError(w, "неверный формат даты", http.StatusBadRequest)
		return
	}

	if task.Repeat != "" {
		next, err := NextDate(now, task.Date, task.Repeat)
		if err != nil {
			sendJSONError(w, "неверный формат правила повторения", http.StatusBadRequest)
			return
		}
		if parsedDate.Format("20060102") < today {
			task.Date = next
		}
	} else if parsedDate.Format("20060102") < today {
		task.Date = today
	}

	if err := db.UpdateTask(&task); err != nil {
		sendJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		sendJSONError(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
