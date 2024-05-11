package main

import (
  "context"
  "database/sql"
  "encoding/json"
  "log/slog"
  "net/http"
  "strconv"
  "strings"
  "time"
)

type Motorcycle struct {
  ID          int                `json:"id"`
  OwnerID     int                `json:"owner_id"`
  PostTitle   string             `json:"post_title"`
  Price       float32            `json:"price"`
  Type        string             `json:"type"`
  Mileage     int64              `json:"mileage"`
  Brand       string             `json:"brand"`
  Model       string             `json:"model"`
  Year        int                `json:"year"`
  Engine      string             `json:"engine"`
  Color       string             `json:"color"`
  Description string             `json:"description"`
  Location    string             `json:"location"`
  Images      []*MotorcycleImage `json:"images"`
  CreatedAt   time.Time          `json:"created_at"`
  UpdatedAt   time.Time          `json:"updated_at"`
}

type MotorcycleImage struct {
  ID        int       `json:"id"`
  URL       string    `json:"url"`
  CreatedAt time.Time `json:"created_at"`
  UpdatedAt time.Time `json:"updated_at"`
}

type MotorcycleCreation struct {
  PostTitle   string  `json:"post_title"`
  Price       float32 `json:"price"`
  Type        string  `json:"type"`
  Mileage     int64   `json:"mileage"`
  Brand       string  `json:"brand"`
  Model       string  `json:"model"`
  Year        int     `json:"year"`
  Engine      string  `json:"engine"`
  Color       string  `json:"color"`
  Description string  `json:"description"`
  Location    string  `json:"location"`
}

type MotorcycleService struct {
  db *sql.DB
}

func NewMotorcycleService(db *sql.DB) *MotorcycleService {
  return &MotorcycleService{db}
}

func (s *MotorcycleService) Create(ctx context.Context, ownerID int, creation *MotorcycleCreation) (insertedID int, err error) {
  tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
  if nil != err {
    slog.Error(err.Error())
    return 0, err
  }

  defer tx.Rollback()

  createMotorcycleQuery := `
  INSERT INTO "motorcycle" (owner_id,
                            post_title,
                            price,
                            type,
                            mileage,
                            brand,
                            model,
                            year,
                            engine,
                            color,
                            description,
                            location)
                    VALUES (@owner_id,
                            @post_title,
                            @price,
                            @type,
                            @mileage,
                            @brand,
                            @model,
                            @year,
                            @engine,
                            @color,
                            @description,
                            @location)
    RETURNING id;`

  ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
  defer cancel()

  err = tx.QueryRowContext(ctx, createMotorcycleQuery,
    sql.Named("owner_id", ownerID),
    sql.Named("post_title", strings.TrimSpace(creation.PostTitle)),
    sql.Named("price", creation.Price),
    sql.Named("type", strings.TrimSpace(creation.Type)),
    sql.Named("mileage", creation.Mileage),
    sql.Named("brand", strings.TrimSpace(creation.Brand)),
    sql.Named("model", strings.TrimSpace(creation.Model)),
    sql.Named("year", creation.Year),
    sql.Named("engine", strings.TrimSpace(creation.Engine)),
    sql.Named("color", strings.TrimSpace(creation.Color)),
    sql.Named("description", strings.TrimSpace(creation.Description)),
    sql.Named("location", strings.TrimSpace(creation.Location))).
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

type MotorcycleHandler struct {
  s *MotorcycleService
}

func NewMotorcycleHandler(service *MotorcycleService) *MotorcycleHandler {
  return &MotorcycleHandler{service}
}

func (h *MotorcycleHandler) Create(w http.ResponseWriter, r *http.Request) {
  ownerID := r.Context().Value("user_id").(int)
  creation := MotorcycleCreation{}

  decoder := json.NewDecoder(r.Body)
  err := decoder.Decode(&creation)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  insertedID, err := h.s.Create(r.Context(), ownerID, &creation)
  if nil != err {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(`{ "inserted_id":` + strconv.Itoa(insertedID) + `}`))
}
