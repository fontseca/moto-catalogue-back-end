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
  CreatedAt   string             `json:"created_at"`
  UpdatedAt   string             `json:"updated_at"`
}

type MotorcycleImage struct {
  ID        int    `json:"id"`
  URL       string `json:"url"`
  CreatedAt string `json:"created_at"`
  UpdatedAt string `json:"updated_at"`
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

func (s *MotorcycleService) GetFromUser(ctx context.Context, ownerID, page int) (motorcycles []*Motorcycle, err error) {
  getMotorcyclesImagesQuery := `
       SELECT mi.motorcycle_id, 
              mi.id,
              mi.url,
              mi.created_at,
              mi.updated_at
       FROM motorcycle_image mi
  LEFT JOIN motorcycle m
         ON m.id = mi.motorcycle_id  
      WHERE m.owner_id = $1;`

  ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  result, err := s.db.QueryContext(ctx, getMotorcyclesImagesQuery, ownerID)
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  defer result.Close()

  motorcycleImagesDictionary := map[int][]*MotorcycleImage{}

  for result.Next() {
    var (
      motorcycleID int
      image        MotorcycleImage
    )

    err = result.Scan(
      &motorcycleID,
      &image.ID,
      &image.URL,
      &image.CreatedAt,
      &image.UpdatedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    motorcycleImagesDictionary[motorcycleID] = append(motorcycleImagesDictionary[motorcycleID], &image)
  }

  getUserMotorcyclesQuery := `
  SELECT id,
         owner_id,
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
         location,
         created_at,
         updated_at
    FROM motorcycle
   WHERE owner_id = @owner_id
   ORDER BY created_at DESC
   LIMIT 10
   OFFSET 10 * (@page - 1);`

  ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
  defer cancel()

  if 0 >= page {
    page = 1
  }

  result, err = s.db.QueryContext(ctx, getUserMotorcyclesQuery,
    sql.Named("owner_id", ownerID), sql.Named("page", page))
  if nil != err {
    slog.Error(err.Error())
    return nil, err
  }

  defer result.Close()

  motorcycles = make([]*Motorcycle, 0)

  for result.Next() {
    var motorcycle Motorcycle

    err = result.Scan(
      &motorcycle.ID,
      &motorcycle.OwnerID,
      &motorcycle.PostTitle,
      &motorcycle.Price,
      &motorcycle.Type,
      &motorcycle.Mileage,
      &motorcycle.Brand,
      &motorcycle.Model,
      &motorcycle.Year,
      &motorcycle.Engine,
      &motorcycle.Color,
      &motorcycle.Description,
      &motorcycle.Location,
      &motorcycle.CreatedAt,
      &motorcycle.UpdatedAt,
    )

    if nil != err {
      slog.Error(err.Error())
      return nil, err
    }

    motorcycle.Images = motorcycleImagesDictionary[motorcycle.ID]
    motorcycles = append(motorcycles, &motorcycle)
  }

  return motorcycles, nil
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

func (h *MotorcycleHandler) Get(w http.ResponseWriter, r *http.Request) {
  ownerID := r.Context().Value("user_id").(int)
  pageStr := r.URL.Query().Get("page")

  page, err := strconv.Atoi(pageStr)
  if nil != err {
    slog.Error(err.Error())
    page = 1
  }

  motorcycles, err := h.s.GetFromUser(r.Context(), ownerID, page)
  if nil != err {
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  response, err := json.Marshal(motorcycles)
  if nil != err {
    slog.Error(err.Error())
    w.WriteHeader(http.StatusInternalServerError)
    return
  }

  w.WriteHeader(http.StatusOK)
  w.Write(response)
}
