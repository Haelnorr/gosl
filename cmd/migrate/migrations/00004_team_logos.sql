-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS team_logo(
    url TEXT NOT NULL,
    team_id INTEGER NOT NULL,
    message_id TEXT NOT NULL,
    channel_id TEXT NOT NULL,
    uploaded INTEGER NOT NULL,
    FOREIGN KEY(team_id) REFERENCES team(id)
) STRICT;
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS team_logo;
-- +goose StatementBegin
-- +goose StatementEnd
