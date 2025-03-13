-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS season(
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    start TEXT,
    reg_season_end TEXT,
    finals_end TEXT,
    active INTEGER NOT NULL DEFAULT 0,
    registration_open INTEGER NOT NULL DEFAULT 0
) STRICT;

CREATE TABLE IF NOT EXISTS league(
    id INTEGER PRIMARY KEY,
    division TEXT NOT NULL,
    season_id TEXT NOT NULL,
    enabled INTEGER NOT NULL DEFAULT 0,
    UNIQUE(division, season_id),
    FOREIGN KEY(season_id) REFERENCES season(id)
) STRICT;

CREATE TABLE IF NOT EXISTS transfer_window(
    id INTEGER PRIMARY KEY,
    season_id TEXT NOT NULL,
    start TEXT,
    end TEXT,
    FOREIGN KEY(season_id) REFERENCES season(id)
) STRICT;

CREATE TABLE IF NOT EXISTS team(
    id INTEGER PRIMARY KEY,
    abbreviation TEXT UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    manager_id INTEGER NOT NULL,
    FOREIGN KEY(manager_id) REFERENCES player(id)
) STRICT;

CREATE TABLE IF NOT EXISTS team_league(
    team_id INTEGER NOT NULL,
    league_id INTEGER NOT NULL,
    PRIMARY KEY(team_id, league_id),
    FOREIGN KEY(team_id) REFERENCES team(id),
    FOREIGN KEY(league_id) REFERENCES league(id)
) STRICT;

CREATE TABLE IF NOT EXISTS player(
    id INTEGER PRIMARY KEY,
    slap_id INTEGER UNIQUE,
    name TEXT UNIQUE,
    discord_id TEXT UNIQUE
) STRICT;

CREATE TABLE IF NOT EXISTS player_team(
    player_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    joined TEXT NOT NULL,
    left TEXT,
    FOREIGN KEY(player_id) REFERENCES player(id),
    FOREIGN KEY(team_id) REFERENCES team(id)
) STRICT;

CREATE TABLE IF NOT EXISTS player_team_invite(
    id INTEGER PRIMARY KEY,
    player_id INTEGER NOT NULL,
    team_id INTEGER NOT NULL,
    status INTEGER,
    approved INTEGER,
    FOREIGN KEY(player_id) REFERENCES player(id),
    FOREIGN KEY(team_id) REFERENCES team(id)
) STRICT;

CREATE TABLE IF NOT EXISTS team_registration(
    id INTEGER PRIMARY KEY,
    team_id INTEGER NOT NULL,
    season_id TEXT NOT NULL,
    preferred_league TEXT NOT NULL,
    approved INTEGER,
    placed INTEGER DEFAULT 0,
    FOREIGN KEY(team_id) REFERENCES team(id),
    FOREIGN KEY(season_id) REFERENCES season(id)
) STRICT;

CREATE TABLE IF NOT EXISTS free_agent(
    player_id INTEGER NOT NULL,
    league_id INTEGER NOT NULL,
    PRIMARY KEY(player_id, league_id),
    FOREIGN KEY(player_id) REFERENCES player(id),
    FOREIGN KEY(league_id) REFERENCES league(id)
) STRICT;

CREATE TABLE IF NOT EXISTS free_agent_registration(
    id INTEGER PRIMARY KEY,
    player_id INTEGER NOT NULL,
    season_id TEXT NOT NULL,
    preferred_league TEXT NOT NULL,
    approved INTEGER,
    placed INTEGER DEFAULT 0,
    FOREIGN KEY(player_id) REFERENCES player(id),
    FOREIGN KEY(season_id) REFERENCES season(id)
) STRICT;

CREATE TRIGGER IF NOT EXISTS enforce_single_active_season
BEFORE UPDATE ON season
FOR EACH ROW WHEN NEW.active = 1
BEGIN
UPDATE season SET active = 0 WHERE id != NEW.id;
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS season;
DROP TABLE IF EXISTS league;
DROP TABLE IF EXISTS transfer_window;
DROP TABLE IF EXISTS team;
DROP TABLE IF EXISTS team_league;
DROP TABLE IF EXISTS player;
DROP TABLE IF EXISTS player_team;
DROP TABLE IF EXISTS player_team_invite;
DROP TABLE IF EXISTS team_registration;
DROP TABLE IF EXISTS free_agent;
DROP TABLE IF EXISTS free_agent_registration;
DROP TRIGGER IF EXISTS enforce_single_active_season;
-- +goose StatementEnd
