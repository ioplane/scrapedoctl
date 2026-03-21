-- +goose Up
CREATE TABLE scrapes (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    request_hash TEXT NOT NULL,
    url          TEXT NOT NULL,
    method       TEXT NOT NULL,
    content      TEXT NOT NULL,
    metadata     TEXT NOT NULL,
    created_at   DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_scrapes_hash ON scrapes(request_hash);
CREATE INDEX idx_scrapes_url  ON scrapes(url);
CREATE INDEX idx_scrapes_date ON scrapes(created_at DESC);

-- +goose Down
DROP TABLE scrapes;
