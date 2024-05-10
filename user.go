package main

import (
  "context"
  "database/sql"
  "encoding/json"
  "golang.org/x/crypto/bcrypt"
  "log/slog"
  "net/http"
  "strconv"
  "time"
)

type User struct {
  ID          int    `json:"id"`
  FirstName   string `json:"first_name"`
  MiddleName  string `json:"middle_name"`
  LastName    string `json:"last_name"`
  Surname     string `json:"surname"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  Password    string
  CreatedAt   string `json:"created_at"`
  UpdatedAt   string `json:"updated_at"`
}

type UserCredentials struct {
  FirstName   string `json:"first_name"`
  MiddleName  string `json:"middle_name"`
  LastName    string `json:"last_name"`
  Surname     string `json:"surname"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  Password    string `json:"password"`
}

type UserService struct {
  db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
  return &UserService{db}
}

func (s *UserService) SignUp(ctx context.Context, credentials *UserCredentials) (insertedID int, err error) {
  tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  defer tx.Rollback()

  registerUserQuery := `
  INSERT INTO "user" (first_name, middle_name, last_name, surname, email, phone_number, password)
              VALUES (@first_name, @middle_name, @last_name, @surname, @email, @phone_number, @password)
    RETURNING id;`

  hashedPassword, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcrypt.MinCost)
  if nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  err = tx.QueryRowContext(ctx, registerUserQuery,
    sql.Named("first_name", credentials.FirstName),
    sql.Named("middle_name", credentials.MiddleName),
    sql.Named("last_name", credentials.LastName),
    sql.Named("surname", credentials.Surname),
    sql.Named("email", credentials.Email),
    sql.Named("phone_number", credentials.PhoneNumber),
    sql.Named("password", hashedPassword)).
    Scan(&insertedID)

  if nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  return insertedID, nil
}

type UserHandler struct {
  s *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
  return &UserHandler{service}
}

func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
  credentials := UserCredentials{}

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&credentials)
  if err != nil {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  insertedID, err := h.s.SignUp(context.TODO(), &credentials)
  if err != nil {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(`{"inserted_id":` + strconv.Itoa(insertedID) + `}`))
}
