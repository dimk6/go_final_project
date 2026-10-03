package api

import (
"database/sql"
"encoding/json"
"net/http"
"strconv"
"time"

"go_final_project/pkg/db"
)

type Task struct {
ID      string `json:"id,omitempty"`
Date    string `json:"date"`
Title   string `json:"title"`
Comment string `json:"comment,omitempty"`
Repeat  string `json:"repeat,omitempty"`
}

type ResponseID struct {
ID int64 `json:"id,omitempty"`
}

type ResponseError struct {
Error string `json:"error,omitempty"`
}

func sendJSONError(w http.ResponseWriter, msg string, code int) {
w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(code)
json.NewEncoder(w).Encode(ResponseError{Error: msg})
}

// TaskHandler маршрутизирует запросы к /api/task (POST, GET, PUT, DELETE)
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
var task Task
if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
return
}

if task.Title == "" {
sendJSONError(w, "не указан заголовок задачи", http.StatusBadRequest)
return
}

now := time.Now()
todayStr := now.Format(DateFormat)

if task.Date == "" {
task.Date = todayStr
}

parsedDate, err := time.Parse(DateFormat, task.Date)
if err != nil {
sendJSONError(w, "неверный формат даты задачи", http.StatusBadRequest)
return
}

nowDate, _ := time.Parse(DateFormat, todayStr)

if parsedDate.Before(nowDate) {
if task.Repeat == "" {
task.Date = todayStr
} else {
next, err := NextDate(nowDate, task.Date, task.Repeat)
if err != nil {
sendJSONError(w, err.Error(), http.StatusBadRequest)
return
}
task.Date = next
}
}

res, err := db.DB.Exec(
"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
task.Date, task.Title, task.Comment, task.Repeat,
)
if err != nil {
sendJSONError(w, "ошибка сохранения задачи в БД", http.StatusInternalServerError)
return
}

id, err := res.LastInsertId()
if err != nil {
sendJSONError(w, "ошибка получения ID задачи", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(ResponseID{ID: id})
}

func getTask(w http.ResponseWriter, r *http.Request) {
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
date    string
title   string
comment string
repeat  string
)

err = db.DB.QueryRow("SELECT date, title, comment, repeat FROM scheduler WHERE id = ?", id).
Scan(&date, &title, &comment, &repeat)
if err == sql.ErrNoRows {
sendJSONError(w, "задача не найдена", http.StatusNotFound)
return
} else if err != nil {
sendJSONError(w, "ошибка чтения задачи", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
json.NewEncoder(w).Encode(Task{
ID:      idStr,
Date:    date,
Title:   title,
Comment: comment,
Repeat:  repeat,
})
}

func updateTask(w http.ResponseWriter, r *http.Request) {
var task Task
if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
sendJSONError(w, "ошибка десериализации JSON", http.StatusBadRequest)
return
}

if task.ID == "" {
sendJSONError(w, "не указан идентификатор задачи", http.StatusBadRequest)
return
}

if task.Title == "" {
sendJSONError(w, "не указан заголовок задачи", http.StatusBadRequest)
return
}

id, err := strconv.ParseInt(task.ID, 10, 64)
if err != nil {
sendJSONError(w, "неверный идентификатор задачи", http.StatusBadRequest)
return
}

now := time.Now()
todayStr := now.Format(DateFormat)

if task.Date == "" {
task.Date = todayStr
}

parsedDate, err := time.Parse(DateFormat, task.Date)
if err != nil {
sendJSONError(w, "неверный формат даты задачи", http.StatusBadRequest)
return
}

nowDate, _ := time.Parse(DateFormat, todayStr)

if parsedDate.Before(nowDate) {
if task.Repeat == "" {
task.Date = todayStr
} else {
next, err := NextDate(nowDate, task.Date, task.Repeat)
if err != nil {
sendJSONError(w, err.Error(), http.StatusBadRequest)
return
}
task.Date = next
}
}

res, err := db.DB.Exec(
"UPDATE scheduler SET date = ?, title = ?, comment = ?, repeat = ? WHERE id = ?",
task.Date, task.Title, task.Comment, task.Repeat, id,
)
if err != nil {
sendJSONError(w, "ошибка обновления задачи в БД", http.StatusInternalServerError)
return
}

rowsAffected, err := res.RowsAffected()
if err != nil || rowsAffected == 0 {
sendJSONError(w, "задача не найдена", http.StatusNotFound)
return
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
w.Write([]byte("{}"))
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
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

_, err = db.DB.Exec("DELETE FROM scheduler WHERE id = ?", id)
if err != nil {
sendJSONError(w, "ошибка удаления задачи из БД", http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json; charset=utf-8")
w.WriteHeader(http.StatusOK)
w.Write([]byte("{}"))
}
