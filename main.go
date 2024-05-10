package main

import (
  "database/sql"
  _ "github.com/mattn/go-sqlite3"
  "log"
  "net"
  "net/http"
  "time"
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

  mux := http.NewServeMux()

  listener, err := net.Listen("tcp", ":3456")
  if nil != err {
    log.Fatalf("could not open listener: %v", err)
  }

  defer listener.Close()

  server := http.Server{
    Addr:              "",
    Handler:           mux,
    ReadTimeout:       5 * time.Minute,
    ReadHeaderTimeout: 5 * time.Minute,
    WriteTimeout:      5 * time.Minute,
    IdleTimeout:       120 * time.Minute,
  }

  log.Fatal(server.Serve(listener))
}
