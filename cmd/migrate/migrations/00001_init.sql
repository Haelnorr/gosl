-- +goose Up
-- +goose StatementBegin
PRAGMA foreign_keys=ON;
CREATE TABLE IF NOT EXISTS jwtblacklist (
jti TEXT PRIMARY KEY CHECK(jti GLOB '[0-9a-fA-F-]*'),
exp INTEGER NOT NULL
) STRICT;
CREATE TABLE IF NOT EXISTS "users" (
    id INTEGER PRIMARY KEY AUTOINCREMENT, 
    username TEXT NOT NULL UNIQUE, 
    password_hash TEXT DEFAULT "", 
    created_at INTEGER DEFAULT (unixepoch()),
    bio TEXT DEFAULT ""
) STRICT;
CREATE TRIGGER IF NOT EXISTS cleanup_expired_tokens
AFTER INSERT ON jwtblacklist
BEGIN
DELETE FROM jwtblacklist WHERE exp < strftime('%s', 'now');
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS cleanup_expired_tokens;
DROP TABLE IF EXISTS jwtblacklist;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd

