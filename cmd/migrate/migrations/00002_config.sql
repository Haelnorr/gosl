-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS config_roles(
    permission INTEGER NOT NULL,
    role_id TEXT NOT NULL,
    PRIMARY KEY(permission, role_id)
) STRICT;

CREATE TABLE IF NOT EXISTS config_channels(
    purpose INTEGER NOT NULL,
    channel_id TEXT NOT NULL,
    PRIMARY KEY(purpose, channel_id)
) STRICT;

CREATE TABLE IF NOT EXISTS config_messages(
    purpose INTEGER PRIMARY KEY,
    message_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    UNIQUE(message_id, channel_id)
) STRICT;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS config_roles;
DROP TABLE IF EXISTS config_channels;
DROP TABLE IF EXISTS config_messages;
-- +goose StatementEnd
