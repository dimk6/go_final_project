package main

import (
"log"
"os"

"go_final_project/pkg/db"
"go_final_project/pkg/server"
)

func main() {
dbFile := "scheduler.db"
if envDB := os.Getenv("TODO_DBFILE"); envDB != "" {
dbFile = envDB
}

if err := db.Init(dbFile); err != nil {
log.Fatalf("Ошибка инициализации БД: %v", err)
}
defer db.DB.Close()

if err := server.Start(); err != nil {
log.Fatalf("Ошибка запуска сервера: %v", err)
}
}
