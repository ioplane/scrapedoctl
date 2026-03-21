-- name: GetLatestScrape :one
SELECT * FROM scrapes
WHERE request_hash = ?
ORDER BY created_at DESC, id DESC
LIMIT 1;

-- name: InsertScrape :one
INSERT INTO scrapes (
    request_hash, url, method, content, metadata
) VALUES (
    ?, ?, ?, ?, ?
) RETURNING *;

-- name: GetHistoryByUrl :many
SELECT * FROM scrapes
WHERE url = ?
ORDER BY created_at DESC, id DESC;

-- name: DeleteOldVersions :exec
DELETE FROM scrapes
WHERE scrapes.request_hash = ?
AND scrapes.id NOT IN (
    SELECT s2.id FROM (
        SELECT s3.id FROM scrapes s3
        WHERE s3.request_hash = ?
        ORDER BY s3.created_at DESC, s3.id DESC
        LIMIT ?
    ) s2
);

-- name: ClearCache :exec
DELETE FROM scrapes;

-- name: GetStats :one
SELECT
    COUNT(*) as total_count,
    SUM(length(content)) as total_size_bytes
FROM scrapes;

-- name: InsertUsage :exec
INSERT INTO usage_log (provider, engine, action, query, url, credits)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetUsageSince :many
SELECT * FROM usage_log
WHERE created_at >= ?
ORDER BY created_at DESC;

-- name: GetUsageByProvider :many
SELECT provider, action,
       COUNT(*) as count,
       SUM(credits) as total_credits
FROM usage_log
WHERE created_at >= ?
GROUP BY provider, action
ORDER BY provider, action;

-- name: GetUsageTotal :one
SELECT COUNT(*) as total_count,
       SUM(credits) as total_credits
FROM usage_log
WHERE created_at >= ?;
