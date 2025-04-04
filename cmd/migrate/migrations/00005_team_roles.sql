-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS team_role(
    team_id INTEGER NOT NULL,
    role_id TEXT NOT NULL,
    UNIQUE(team_id),
    FOREIGN KEY(team_id) REFERENCES team(id)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS team_role;
-- +goose StatementEnd
