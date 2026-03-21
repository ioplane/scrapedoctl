-- +goose Up
CREATE TABLE usage_log (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    provider   TEXT NOT NULL,
    engine     TEXT NOT NULL DEFAULT '',
    action     TEXT NOT NULL,
    query      TEXT NOT NULL DEFAULT '',
    url        TEXT NOT NULL DEFAULT '',
    credits    INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_usage_provider ON usage_log(provider);
CREATE INDEX idx_usage_action ON usage_log(action);
CREATE INDEX idx_usage_date ON usage_log(created_at DESC);

-- +goose Down
DROP TABLE usage_log;
