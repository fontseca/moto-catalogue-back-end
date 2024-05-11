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
  "strings"
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

type UserUpdate struct {
  FirstName   string `json:"first_name"`
  MiddleName  string `json:"middle_name"`
  LastName    string `json:"last_name"`
  Surname     string `json:"surname"`
  Email       string `json:"email"`
  PhoneNumber string `json:"phone_number"`
  Password    string `json:"password"`
  PictureURL  string `json:"picture_url"`
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
    sql.Named("first_name", strings.TrimSpace(credentials.FirstName)),
    sql.Named("middle_name", strings.TrimSpace(credentials.MiddleName)),
    sql.Named("last_name", strings.TrimSpace(credentials.LastName)),
    sql.Named("surname", strings.TrimSpace(credentials.Surname)),
    sql.Named("email", strings.TrimSpace(credentials.Email)),
    sql.Named("phone_number", strings.TrimSpace(credentials.PhoneNumber)),
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
    userID        int
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

func (s *UserService) Get(ctx context.Context, page int) (users []*User, err error) {
  getUsersQuery := `
  SELECT id,
         first_name,
         middle_name,
         last_name,
         surname,
         email,
         phone_number,
         picture_url,
         created_at,
         updated_at
    FROM "user"
ORDER BY created_at
   LIMIT 10
   OFFSET 10 * (@page - 1);`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := s.db.QueryContext(ctx, getUsersQuery, sql.Named("page", page))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  users = make([]*User, 0)

  for result.Next() {
    user := new(User)

    err = result.Scan(
      &user.ID,
      &user.FirstName,
      &user.MiddleName,
      &user.LastName,
      &user.Surname,
      &user.Email,
      &user.PhoneNumber,
      &user.PictureURL,
      &user.CreatedAt,
      &user.UpdatedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    users = append(users, user)
  }

  return users, nil
}

func (s *UserService) Update(ctx context.Context, id int, update *UserUpdate) error {
  tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return err
  }

  defer tx.Rollback()

  updateUserQuery := `
  UPDATE "user"
     SET first_name = coalesce(nullif(@first_name, ''), first_name),
         middle_name = coalesce(nullif(@middle_name, ''), middle_name),
         last_name = coalesce(nullif(@last_name, ''), last_name),
         surname = coalesce(nullif(@surname, ''), surname),
         email = coalesce(nullif(@email, ''), email),
         phone_number = coalesce(nullif(@phone_number, ''), phone_number),
         picture_url = coalesce(nullif(@picture_url, ''), picture_url),
         password = coalesce(nullif(@password, ''), password),
         updated_at = current_timestamp
   WHERE id = @id;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  result, err := tx.ExecContext(ctx, updateUserQuery,
    sql.Named("id", id),
    sql.Named("first_name", strings.TrimSpace(update.FirstName)),
    sql.Named("middle_name", strings.TrimSpace(update.MiddleName)),
    sql.Named("last_name", strings.TrimSpace(update.LastName)),
    sql.Named("surname", strings.TrimSpace(update.Surname)),
    sql.Named("email", strings.TrimSpace(update.Email)),
    sql.Named("phone_number", strings.TrimSpace(update.PhoneNumber)),
    sql.Named("picture_url", strings.TrimSpace(update.PictureURL)),
    sql.Named("password", strings.TrimSpace(update.Password)),
  )

  if nil != err {
    slog.Error(err.Error())
    return err
  }

  if affected, _ := result.RowsAffected(); 1 != affected {
    return errors.New("user not found")
  }

  if err = tx.Commit(); nil != err {
    slog.Error(err.Error())
    return err
  }

  return nil
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

  token, err := h.s.SignIn(r.Context(), &credentials)
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

  user, err := h.s.GetByID(r.Context(), userID)
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

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
  userID := r.Context().Value("user_id").(int)

  user, err := h.s.GetByID(r.Context(), userID)
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

func (h *UserHandler) Get(w http.ResponseWriter, r *http.Request) {
  pageStr := r.URL.Query().Get("page")
  page := 1

  var err error

  if "" != pageStr {
    page, err = strconv.Atoi(pageStr)
    if nil != err {
      slog.Error(err.Error())
      w.WriteHeader(http.StatusInternalServerError)
      return
    }
  }

  users, err := h.s.Get(r.Context(), page)
  if nil != err {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  response, err := json.Marshal(users)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(response)
}

func (h *UserHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
  userID := r.Context().Value("user_id").(int)
  userUpdate := UserUpdate{}

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&userUpdate)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  err = h.s.Update(r.Context(), userID, &userUpdate)
  if nil != err {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusNoContent)
}
