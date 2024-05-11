package main

import (
  "context"
  "database/sql"
  "encoding/json"
  "errors"
  "github.com/golang-jwt/jwt/v5"
  "golang.org/x/crypto/bcrypt"
  "log/slog"
  "net/http"
  "os"
  "strconv"
  "time"
)

type User struct {
  ID          int     `json:"id"`
  FirstName   string  `json:"first_name"`
  MiddleName  *string `json:"middle_name"`
  LastName    *string `json:"last_name"`
  Surname     *string `json:"surname"`
  Email       string  `json:"email"`
  PhoneNumber string  `json:"phone_number"`
  PictureURL  *string `json:"picture_url"`
  Password    string  `json:"-"`
  CreatedAt   string  `json:"created_at"`
  UpdatedAt   string  `json:"updated_at"`
}

type UserCreation struct {
  FirstName   string `json:"first_name"`
  MiddleName  string `json:"middle_name"`
  LastName    string `json:"last_name"`
  Surname     string `json:"surname"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  Password    string `json:"password"`
}

type UserCredentials struct {
  Email    string `json:"email"`
  Password string `json:"password"`
}

type UserService struct {
  db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
  return &UserService{db}
}

func (s *UserService) SignUp(ctx context.Context, credentials *UserCreation) (insertedID int, err error) {
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

func (s *UserService) SignIn(ctx context.Context, credentials *UserCredentials) (token string, err error) {
  getUserPasswordQuery := `
  SELECT id, password
    FROM "user"
   WHERE email = @email;`

  var (
    userID        string
    savedPassword string
  )

  err = s.db.QueryRowContext(ctx, getUserPasswordQuery, sql.Named("email", credentials.Email)).Scan(&userID, &savedPassword)
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  err = bcrypt.CompareHashAndPassword([]byte(savedPassword), []byte(credentials.Password))
  if nil != err {
    if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
      slog.Error(err.Error())
    }

    return "", err
  }

  claims := jwt.MapClaims{
    "iss":     "noda",
    "sub":     "authentication",
    "iat":     jwt.NewNumericDate(time.Now()),
    "exp":     jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
    "user_id": userID,
  }

  t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
  secret := os.Getenv("JWT_SECRET")
  if "" == secret {
    secret = "default secret"
  }

  ss, err := t.SignedString([]byte(secret))
  if nil != err {
    slog.Error(err.Error())
    return "", err
  }

  return ss, nil
}

func (s *UserService) GetByID(ctx context.Context, id int) (user *User, err error) {
  getUserQuery := `
  SELECT id,
         first_name,
         middle_name,
         last_name,
         surname,
         email,
         phone_number,
         picture_url,
         password,
         created_at,
         updated_at
    FROM "user"
   WHERE id = $1;`

  ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
  defer cancel()

  row := s.db.QueryRowContext(ctx, getUserQuery, id)

  user = new(User)

  err = row.Scan(
    &user.ID,
    &user.FirstName,
    &user.MiddleName,
    &user.LastName,
    &user.Surname,
    &user.Email,
    &user.PhoneNumber,
    &user.PictureURL,
    &user.Password,
    &user.CreatedAt,
    &user.UpdatedAt,
  )

  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  return user, nil
}

type UserHandler struct {
  s *UserService
}

func NewUserHandler(service *UserService) *UserHandler {
  return &UserHandler{service}
}

func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
  userCreation := UserCreation{}

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&userCreation)
  if err != nil {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  insertedID, err := h.s.SignUp(context.TODO(), &userCreation)
  if err != nil {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(`{"inserted_id":` + strconv.Itoa(insertedID) + `}`))
}

func (h *UserHandler) SignIn(w http.ResponseWriter, r *http.Request) {
  credentials := UserCredentials{}

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&credentials)
  if err != nil {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  token, err := h.s.SignIn(context.TODO(), &credentials)
  if err != nil {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(`{"token":"` + token + `"}`))
}

func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
  userIDStr := r.PathValue("user_id")

  userID, err := strconv.Atoi(userIDStr)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  user, err := h.s.GetByID(context.TODO(), userID)
  if nil != err {
    if errors.Is(err, sql.ErrNoRows) {
      w.WriteHeader(http.StatusNotFound)
    } else {
      w.WriteHeader(http.StatusInternalServerError)
    }

    return
  }

  response, err := json.Marshal(user)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(response)
}
