-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "config_roles" (
    permission INTEGER NOT NULL, 
    role_id TEXT NOT NULL,
    PRIMARY KEY(permission, role_id)
) STRICT;
CREATE TABLE IF NOT EXISTS "config_channels" (
    purpose INTEGER NOT NULL,
    channel_id TEXT NOT NULL,
    PRIMARY KEY(purpose, channel_id)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS "config_roles";
DROP TABLE IF EXISTS "config_channels";
-- +goose StatementEnd
