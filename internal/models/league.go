package models

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Model of the league table in the database
// Each record in this table covers a single league for a given season
type League struct {
	ID       uint16 // unique ID
	Division string // division of the league i.e. Open, IM, Pro
	SeasonID string // FK -> Season.ID
}

func AddLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	division string,
) error {
	if division != "Open" && division != "IM" && division != "Pro" {
		return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	query := `
INSERT INTO league (division, season_id, enabled)
VALUES (?, ?, 1)
ON CONFLICT (division, season_id)
DO UPDATE SET enabled = 1;
`
	_, err := tx.Exec(ctx, query, division, seasonID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func GetLeagues(
	ctx context.Context,
	tx db.SafeTX,
	seasonID string,
	enabledOnly bool,
) (*[]League, error) {
	enabledMod := ""
	if enabledOnly {
		enabledMod = "AND enabled = 1"
	}
	query := `SELECT id, division FROM league WHERE season_id = ? COLLATE NOCASE %s;`
	rows, err := tx.Query(ctx, fmt.Sprintf(query, enabledMod), seasonID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var leagues []League
	for rows.Next() {
		var league League
		err = rows.Scan(&league.ID, &league.Division)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		league.SeasonID = seasonID
		leagues = append(leagues, league)
	}
	return &leagues, nil
}

func RemoveLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	division string,
) error {
	if division != "Open" && division != "IM" && division != "Pro" {
		return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	query := `UPDATE league SET enabled = 0 WHERE division = ? AND season_id = ?;`
	_, err := tx.Exec(ctx, query, division, seasonID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func SetLeagues(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	divisions []string,
) error {
	allDivs := map[string]bool{
		"Open": false,
		"IM":   false,
		"Pro":  false,
	}
	for _, division := range divisions {
		if division != "Open" && division != "IM" && division != "Pro" {
			return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
		}
		allDivs[division] = true
	}
	for division, enabled := range allDivs {
		if enabled {
			err := AddLeague(ctx, tx, seasonID, division)
			if err != nil {
				return errors.Wrap(err, "AddLeague")
			}
		} else {
			err := RemoveLeague(ctx, tx, seasonID, division)
			if err != nil {
				return errors.Wrap(err, "RemoveLeague")
			}
		}
	}
	return nil
}

func (l *League) GetTeams(ctx context.Context, tx db.SafeTX) (*[]Team, error) {
	query := `
SELECT t.id, t.abbreviation, t.name, t.manager_id, p.name, t.color 
FROM team t 
JOIN team_league tl ON tl.team_id = t.id
JOIN player p ON t.manager_id = p.id
WHERE tl.league_id = ?;`
	rows, err := tx.Query(ctx, query, l.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var teams []Team
	for rows.Next() {
		var team Team
		var color string
		err = rows.Scan(&team.ID, &team.Abbreviation, &team.Name,
			&team.ManagerID, &team.ManagerName, &color)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Wrap(err, "rows.Scan")
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

func (l *League) GetFreeAgents(
	ctx context.Context,
	tx db.SafeTX,
) (*[]Player, error) {
	query := `
SELECT p.id, p.slap_id, p.name, p.discord_id
FROM player p 
JOIN free_agent fa ON fa.player_id = p.id
WHERE fa.league_id = ?;
`
	rows, err := tx.Query(ctx, query, l.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var players []Player
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
