package api

import (
	"encoding/json"
	"net/http"
	"time"

	"go_final_project/pkg/db"
)

type TasksResponse struct {
	Tasks []db.Task `json:"tasks"`
}

func TasksHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		return
	}

	search := r.FormValue("search")
	limit := 50

	var tasks []db.Task
	var err error

	if search != "" {
		if t, parseErr := time.Parse("02.01.2006", search); parseErr == nil {
			tasks, err = db.SearchTasksByDate(t.Format("20060102"), limit)
		} else {
			tasks, err = db.SearchTasksByText(search, limit)
		}
	} else {
		tasks, err = db.GetTasks(limit)
	}

	if err != nil {
		sendJSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(TasksResponse{Tasks: tasks})
}
