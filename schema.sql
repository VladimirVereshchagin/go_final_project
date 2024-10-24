CREATE TABLE IF NOT EXISTS scheduler (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    date TEXT NOT NULL,
    title TEXT NOT NULL,
    comment TEXT,
    repeat TEXT DEFAULT '' NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_date ON scheduler(date);