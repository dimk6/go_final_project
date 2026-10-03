package api

import (
"encoding/json"
"net/http"
"strconv"
"time"

"go_final_project/pkg/db"
)

type TasksResponse struct {
Tasks []Task `json:"tasks"`
}

// TasksHandler обрабатывает GET /api/tasks
func TasksHandler(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
sendJSONError(w, "метод не поддерживается", http.StatusMethodNotAllowed)
return
}

search := r.URL.Query().Get("search")
const limit = 50

var query string
var args []interface{}

if search != "" {
// Проверяем, является ли search датой в формате 02.01.2006
if parsedDate, err := time.Parse("02.01.2006", search); err == nil {
query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date ASC LIMIT ?"
args = append(args, parsedDate.Format(DateFormat), limit)
} else {
// Иначе ищем подстроку в title или comment
pattern := "%" + search + "%"
query = "SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date ASC LIMIT ?"
args = append(args, pattern, pattern, limit)
}
} else {
query = "SELECT id, date, title, comment, repeat FROM scheduler ORDER BY date ASC LIMIT ?"
args = append(args, limit)
}

rows, err := db.DB.Query(query, args...)
if err != nil {
sendJSONError(w, "ошибка получения задач из БД", http.StatusInternalServerError)
return
}
defer rows.Close()

tasks := make([]Task, 0)
for rows.Next() {
var (
idInt   int64
date    string
title   string
comment string
repeat  string
)
if err := rows.Scan(&idInt, &date, &title, &comment, &repeat); err != nil {
sendJSONError(w, "ошибка чтения строки задачи", http.StatusInternalServerError)
return
}
tasks = append(tasks, Task{
ID:      strconv.FormatInt(idInt, 10),
Date:    date,
Title:   title,
Comment: comment,
Repeat:  repeat,
})
}

if err := rows.Err(); err != nil {
sendJSONError(w, "ошибка обхода результатов", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(TasksResponse{Tasks: tasks})
}
