package main

import (
  "context"
  "database/sql"
  "github.com/golang-jwt/jwt/v5"
  _ "github.com/mattn/go-sqlite3"
  "log"
  "log/slog"
  "net"
  "net/http"
  "os"
  "strings"
  "time"
)

func setHeader(key, value string) func(http.Handler) http.Handler {
  return func(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
      w.Header().Add(key, value)
      next.ServeHTTP(w, r)
    })
  }
}

func withAuthorization(next http.HandlerFunc) http.HandlerFunc {
  secret := os.Getenv("JWT_SECRET")
  if "" == secret {
    secret = "default secret"
  }

  return func(w http.ResponseWriter, r *http.Request) {
    authorization := r.Header.Get("Authorization")
    if "" == authorization {
      w.Header().Set("WWW-Authenticate", "Bearer realm=\"access to system\"")
      w.WriteHeader(http.StatusUnauthorized)
      return
    }

    tokenStr := strings.Split(authorization, " ")[1]
    token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) { return []byte(secret), nil })
    if nil != err {
      slog.Error(err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    if !token.Valid {
      w.WriteHeader(http.StatusInternalServerError)
      return
    }

    claims := token.Claims.(jwt.MapClaims)
    userID := claims["user_id"].(float64)

    ctx := context.WithValue(r.Context(), "user_id", int(userID))
    r = r.Clone(ctx)

    next.ServeHTTP(w, r)
  }
}

func with(mux *http.ServeMux, middlewares ...func(http.Handler) http.Handler) http.Handler {
  var h http.Handler = mux

  for _, middleware := range middlewares {
    if nil != middleware {
      h = middleware(h)
    }
  }

  return h
}

func main() {
  log.SetFlags(log.LstdFlags | log.Lshortfile)

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

  h := with(mux, setHeader("Content-Type", "application/json"))

  server := http.Server{
    Addr:              "",
    Handler:           h,
    ReadTimeout:       5 * time.Minute,
    ReadHeaderTimeout: 5 * time.Minute,
    WriteTimeout:      5 * time.Minute,
    IdleTimeout:       120 * time.Minute,
  }

  log.Fatal(server.Serve(listener))
}
