CREATE TABLE IF NOT EXISTS "user"
(
  "id"           INTEGER            NOT NULL PRIMARY KEY AUTOINCREMENT,
  "first_name"   VARCHAR(64)        NOT NULL,
  "middle_name"  VARCHAR(64)                 DEFAULT NULL,
  "last_name"    VARCHAR(64)                 DEFAULT NULL,
  "surname"      VARCHAR(64)                 DEFAULT NULL,
  "email"        VARCHAR(240)       NOT NULL UNIQUE,
  "phone_number" VARCHAR(64) UNIQUE NOT NULL,
  "picture_url"  VARCHAR(2048)               DEFAULT NULL,
  "password"     VARCHAR(256)       NOT NULL,
  "created_at"   timestamptz        NOT NULL DEFAULT current_timestamp,
  "updated_at"   timestamptz        NOT NULL DEFAULT current_timestamp
);


CREATE TABLE IF NOT EXISTS "motorcycle"
(
  "id"          INTEGER      NOT NULL PRIMARY KEY AUTOINCREMENT,
  "owner_id"    INTEGER      NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
  "post_title"  VARCHAR(512) NOT NULL DEFAULT 'Untitled',
  "price"       FLOAT        NOT NULL DEFAULT 0.0,
  "type"        VARCHAR(32)  NOT NULL,
  "mileage"     INTEGER      NOT NULL DEFAULT 0,
  "brand"       VARCHAR(128) NOT NULL DEFAULT 'Unknown',
  "model"       VARCHAR(128) NOT NULL DEFAULT 'Unknown',
  "year"        INT          NOT NULL,
  "engine"      VARCHAR(128) NOT NULL DEFAULT 'Unknown',
  "color"       VARCHAR(32)  NOT NULL,
  "description" VARCHAR(512) NOT NULL DEFAULT 'No description',
  "location"    VARCHAR(512) NOT NULL DEFAULT 'Unknown',
  "created_at"  timestamptz  NOT NULL DEFAULT current_timestamp,
  "updated_at"  timestamptz  NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS "motorcycle_image"
(
  "id"            INTEGER       NOT NULL PRIMARY KEY AUTOINCREMENT,
  "motorcycle_id" INTEGER       NOT NULL REFERENCES "motorcycle" ("id") ON DELETE CASCADE,
  "url"           VARCHAR(2048) NOT NULL,
  "created_at"    timestamptz   NOT NULL DEFAULT current_timestamp,
  "updated_at"    timestamptz   NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS "favorite"
(
  "user_id"       INTEGER NOT NULL REFERENCES "user" ("id") ON DELETE CASCADE,
  "motorcycle_id" INTEGER NOT NULL REFERENCES "motorcycle" ("id") ON DELETE CASCADE
);
