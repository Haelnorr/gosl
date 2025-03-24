package models

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/pkg/db"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Model of the player table in the database
// Each row represents a single player
type Player struct {
	ID        uint16 // unique ID
	SlapID    uint32 // unique slapshot player ID
	Name      string // unique player name
	DiscordID string // unique discord ID
}

func GetPlayerByID(
	ctx context.Context,
	tx db.SafeTX,
	playerID uint16,
) (*Player, error) {
	query := `SELECT id, slap_id, name, discord_id FROM player WHERE id = ?;`
	row, err := tx.QueryRow(ctx, query, playerID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var player Player
	err = row.Scan(&player.ID, &player.SlapID, &player.Name, &player.DiscordID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	return &player, nil
}

// Searches the database for a player linked to the given discord ID. If none found
// returns nil
func GetPlayerByDiscordID(
	ctx context.Context,
	tx db.SafeTX,
	discordID string,
) (*Player, error) {
	query := `SELECT id, slap_id, name FROM player WHERE discord_id = ?;`
	row, err := tx.QueryRow(ctx, query, discordID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var player Player
	err = row.Scan(&player.ID, &player.SlapID, &player.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	player.DiscordID = discordID
	return &player, nil
}

// Searches the database for a player linked to the given discord ID. If none found
// returns nil
func GetPlayerBySlapID(
	ctx context.Context,
	tx db.SafeTX,
	slapID uint32,
) (*Player, error) {
	query := `SELECT id, discord_id, name FROM player WHERE slap_id = ?;`
	row, err := tx.QueryRow(ctx, query, slapID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var player Player
	err = row.Scan(&player.ID, &player.DiscordID, &player.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	player.SlapID = slapID
	return &player, nil
}

func CreatePlayer(
	ctx context.Context,
	tx *db.SafeWTX,
	slapID uint32,
	discordID string,
	name string,
) error {
	query := `INSERT INTO player (slap_id, discord_id, name) VALUES (?,?,?);`
	_, err := tx.Exec(ctx, query, slapID, discordID, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			if strings.Contains(err.Error(), "player.name") {
				msg := fmt.Sprintf("Display name must be unique: '%s' is taken", name)
				return errors.New(msg)
			}
		}
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func GetInviteablePlayers(
	ctx context.Context,
	tx db.SafeTX,
	teamID uint16,
) (*[]Player, error) {
	query := `
SELECT p.id, p.slap_id, p.name, p.discord_id FROM player p
LEFT JOIN player_team pt ON pt.player_id = p.id AND pt.left IS NULL
WHERE pt.player_id IS NULL
AND NOT EXISTS (
    SELECT 1 FROM player_team_invite pti
    WHERE pti.player_id = p.id AND pti.team_id = ?
    AND (
        (pti.status IS NULL AND pti.approved IS NULL) OR
        (pti.status IS NULL AND pti.approved = 1) OR
        (pti.status = 1 AND pti.approved IS NULL)
    )
);
`
	rows, err := tx.Query(ctx, query, teamID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	players := []Player{}
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.ID, &player.SlapID, &player.Name, &player.DiscordID)
		if err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}
		players = append(players, player)
	}
	return &players, nil
}

func (p *Player) UpdateDiscordID(
	ctx context.Context,
	tx *db.SafeWTX,
	discordID string,
) error {
	query := `UPDATE player SET discord_id = ? WHERE slap_id = ?;`
	_, err := tx.Exec(ctx, query, discordID, p.SlapID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	p.DiscordID = discordID
	return nil
}

func (p *Player) CurrentTeam(
	ctx context.Context,
	tx db.SafeTX,
) (*PlayerTeam, error) {
	query := `
SELECT pt.team_id, t.name, pt.player_id, p.name, t.manager_id, pt.joined, pt.left
FROM player_team pt
JOIN team t ON pt.team_id = t.id
JOIN player p ON pt.player_id = p.id
WHERE pt.player_id = ? AND pt.joined < ? AND pt.left IS NULL;`
	now := time.Now()
	row, err := tx.QueryRow(ctx, query, p.ID, formatISO8601(&now))
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var team PlayerTeam
	var joined string
	var left sql.NullString
	err = row.Scan(&team.TeamID, &team.TeamName, &team.PlayerID, &team.PlayerName,
		&team.ManagerID, &joined, &left)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	joinedParsed := parseISO8601(&joined)
	team.Joined = *joinedParsed
	if left.Valid {
		team.Left = parseISO8601(&left.String)
	}
	return &team, nil
}

func (p *Player) JoinTeam(
	ctx context.Context,
	tx *db.SafeWTX,
	teamid uint16,
) error {
	currentTeam, err := p.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "p.CurrentTeam")
	}
	if currentTeam != nil {
		return errors.New("Player currently on a team")
	}
	team, err := GetTeamByID(ctx, tx, teamid)
	if err != nil {
		return errors.Wrap(err, "GetTeamByID")
	}
	if team == nil {
		return errors.New("Team does not exist")
	}
	query := `INSERT INTO player_team (player_id, team_id, joined) VALUES (?,?,?);`
	now := time.Now()
	_, err = tx.Exec(ctx, query, p.ID, teamid, formatISO8601(&now))
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func (p *Player) LeaveTeam(
	ctx context.Context,
	tx *db.SafeWTX,
	teamid uint16,
) error {
	currentTeam, err := p.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "p.CurrentTeam")
	}
	if currentTeam == nil {
		return errors.New("Player not on a team")
	}
	team, err := GetTeamByID(ctx, tx, teamid)
	if err != nil {
		return errors.Wrap(err, "GetTeamByID")
	}
	if team == nil {
		return errors.New("Team does not exist")
	}
	if team.ID != currentTeam.TeamID {
		return errors.New("Player is not on that team!")
	}
	query := `
UPDATE player_team SET left = ?
WHERE team_id = ? AND player_id = ? AND left IS NULL;
    `
	now := time.Now()
	_, err = tx.Exec(ctx, query, formatISO8601(&now), team.ID, p.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func (p *Player) GetManagedTeams(
	ctx context.Context,
	tx db.SafeTX,
) (*[]Team, error) {
	query := `
SELECT id, abbreviation, name, manager_id, color 
FROM team WHERE manager_id = ?;`
	rows, err := tx.Query(ctx, query, p.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	teams := []Team{}
	for rows.Next() {
		var team Team
		var color string
		err = rows.Scan(&team.ID, &team.Abbreviation, &team.Name, &team.ManagerID, &color)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Wrap(err, "row.Scan")
		}
		colorint, err := hexToInt(color)
		if err != nil {
			team.Color = 0x181825
		} else {
			team.Color = colorint
		}
		teams = append(teams, team)
	}
	return &teams, nil
}
