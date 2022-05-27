CREATE TABLE fridges(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL,
    min_temp REAL NOT NULL,
    max_temp REAL NOT NULL,
    alerts_enabled INTEGER NOT NULL DEFAULT 0
) STRICT;

CREATE TABLE temperatures(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    value REAL NOT NULL,
    fridge_id INTEGER NOT NULL REFERENCES fridges(id),
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
) STRICT;

CREATE INDEX idx_temperatures_fridge_id_created_at ON temperatures(fridge_id, created_at);
