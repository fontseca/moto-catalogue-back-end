package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "log"
)

func main() {
  db, err := sql.Open("sqlite3", "db.sqlite")
  if nil != err {
    log.Fatalf("could not open database: %v", err)
  }

  defer db.Close()

  err = db.Ping()
  if nil != err {
    log.Fatalf("could not ping database: %v", err)
  }

}
