package server

import (
"log"
"net/http"
"os"

"go_final_project/pkg/api"
)

func Start() error {
port := "7540"
if envPort := os.Getenv("TODO_PORT"); envPort != "" {
port = envPort
}

http.HandleFunc("/api/nextdate", api.NextDateHandler)
http.HandleFunc("/api/signin", api.SignInHandler)

http.HandleFunc("/api/task", api.AuthMiddleware(api.TaskHandler))
http.HandleFunc("/api/tasks", api.AuthMiddleware(api.TasksHandler))
http.HandleFunc("/api/task/done", api.AuthMiddleware(api.TaskDoneHandler))

webDir := "./web"
fileServer := http.FileServer(http.Dir(webDir))
http.Handle("/", fileServer)

addr := "0.0.0.0:" + port
log.Printf("Сервер запущен на %s\n", addr)
return http.ListenAndServe(addr, nil)
}
