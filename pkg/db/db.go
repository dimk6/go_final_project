package db

import (
"database/sql"

_ "modernc.org/sqlite"
)

var DB *sql.DB

const schema = `
CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date CHAR(8) NOT NULL DEFAULT "",
    title VARCHAR(256) NOT NULL DEFAULT "",
    comment TEXT NOT NULL DEFAULT "",
    repeat VARCHAR(128) NOT NULL DEFAULT ""
);

CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler (date);
`

func Init(dbFile string) error {
database, err := sql.Open("sqlite", dbFile)
if err != nil {
return err
}

if err := database.Ping(); err != nil {
return err
}

// Всегда применяем схему (IF NOT EXISTS гарантирует безопасность)
if _, err := database.Exec(schema); err != nil {
return err
}

DB = database
return nil
}
