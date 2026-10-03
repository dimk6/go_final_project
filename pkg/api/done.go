package api

import (
"database/sql"
"net/http"
"strconv"
"time"

"go_final_project/pkg/db"
)

// TaskDoneHandler обрабатывает POST /api/task/done
func TaskDoneHandler(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
return
}

idStr := r.URL.Query().Get("id")
if idStr == "" {
sendJSONError(w, "не указан идентификатор", http.StatusBadRequest)
return
}

id, err := strconv.ParseInt(idStr, 10, 64)
if err != nil {
sendJSONError(w, "неверный идентификатор задачи", http.StatusBadRequest)
return
}

var (
date   string
repeat string
)

err = db.DB.QueryRow("SELECT date, repeat FROM scheduler WHERE id = ?", id).Scan(&date, &repeat)
if err == sql.ErrNoRows {
sendJSONError(w, "задача не найдена", http.StatusNotFound)
return
} else if err != nil {
sendJSONError(w, "ошибка чтения задачи", http.StatusInternalServerError)
return
}

if repeat == "" {
// Обычная задача — удаляем
_, err = db.DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
if err != nil {
sendJSONError(w, "ошибка удаления задачи", http.StatusInternalServerError)
return
}
} else {
// Повторяющаяся задача — вычисляем следующий дедлайн
now := time.Now()
nextDate, err := NextDate(now, date, repeat)
if err != nil {
sendJSONError(w, err.Error(), http.StatusBadRequest)
return
}

_, err = db.DB.Exec("UPDATE scheduler SET date = ? WHERE id = ?", nextDate, id)
if err != nil {
sendJSONError(w, "ошибка обновления даты задачи", http.StatusInternalServerError)
return
}
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
w.Write([]byte("{}"))
}
