package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"
	"regexp"
	"time"

	"github.com/pkg/errors"
)

// Model of the team table in the database
// Each row represents a single team
type Team struct {
	ID           uint16 // unique ID
	Abbreviation string // unique abbreviation of team name
	Name         string // unique team name
	ManagerID    uint16 // FK -> Player.ID
	Color        int    // colour hex
}

func GetTeamByName(
	ctx context.Context,
	tx db.SafeTX,
	name string,
) (*Team, error) {
	query := `
SELECT id, abbreviation, name, manager_id, color 
FROM team WHERE name = ? COLLATE NOCASE;`
	row, err := tx.QueryRow(ctx, query, name)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var team Team
	var color string
	err = row.Scan(&team.ID, &team.Abbreviation, &team.Name, &team.ManagerID, &color)
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
	return &team, nil
}

func GetTeamByAbbr(
	ctx context.Context,
	tx db.SafeTX,
	abbr string,
) (*Team, error) {
	query := `
SELECT id, abbreviation, name, manager_id, color 
FROM team WHERE abbreviation = ? COLLATE NOCASE;`
	row, err := tx.QueryRow(ctx, query, abbr)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var team Team
	var color string
	err = row.Scan(&team.ID, &team.Abbreviation, &team.Name, &team.ManagerID, &color)
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
	return &team, nil
}

func GetTeamByID(
	ctx context.Context,
	tx db.SafeTX,
	id uint16,
) (*Team, error) {
	query := `
SELECT id, abbreviation, name, manager_id, color 
FROM team WHERE id = ?;`
	row, err := tx.QueryRow(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var team Team
	var color string
	err = row.Scan(&team.ID, &team.Abbreviation, &team.Name, &team.ManagerID, &color)
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
	return &team, nil
}

func CreateTeam(
	ctx context.Context,
	tx *db.SafeWTX,
	name string,
	abbr string,
	managerid uint16,
) (*Team, error) {
	query := `
INSERT INTO team (name, abbreviation, manager_id)
VALUES (?,?,?);`
	res, err := tx.Exec(ctx, query, name, abbr, managerid)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Exec")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "res.LastInsertId")
	}
	team, err := GetTeamByID(ctx, tx, uint16(id))
	if err != nil {
		return nil, errors.Wrap(err, "GetTeamByID")
	}
	return team, nil
}

// Get array of all the players for the team. If windowStart or windowEnd are
// provided, list will only contain players who were on the team during the
// specified period. One, neither or both of these filters can be provided.
// If time.Now() is provided in both, will return current players.
func (t *Team) Players(
	ctx context.Context,
	tx db.SafeTX,
	windowStart *time.Time,
	windowEnd *time.Time,
) (*[]Player, error) {
	query := `
SELECT p.id, p.slap_id, p.name, p.discord_id FROM player p
JOIN player_team pt ON p.id = pt.player_id
WHERE pt.team_id = ?
AND (
    pt.joined <= ? 
    AND
    (pt.left IS NULL OR pt.left > ?)
);
`
	if windowStart == nil {
		unixStart := time.Unix(0, 0)
		windowStart = &unixStart
	}
	if windowEnd == nil {
		now := time.Now()
		windowEnd = &now
	}
	rows, err := tx.Query(ctx, query, t.ID, formatISO8601(windowEnd), formatISO8601(windowStart))
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	players := []Player{}
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.ID, &player.SlapID, &player.Name, &player.DiscordID)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		players = append(players, player)
	}
	return &players, nil
}

func (t *Team) Disband(ctx context.Context, tx *db.SafeWTX) error {
	// TODO: Check with LC's conditions for disbanding a team and how to handle
	// in the meantime, just block disband if team placed into a league
	var exists int
	query := `
SELECT EXISTS (
    SELECT 1 FROM team_league tl
    JOIN league l ON tl.league_id = l.id
    JOIN season s ON l.season_id = s.id
    WHERE tl.team_id = ?
    AND l.enabled = 1
    AND s.active = 1
);`
	row, err := tx.QueryRow(ctx, query, t.ID)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&exists)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	if exists == 1 {
		return errors.New("Team cannot be disbanded as they are in an active league")
	}
	query = `
UPDATE player_team SET left = ?
WHERE team_id = ? AND left IS NULL;
    `
	now := time.Now()
	_, err = tx.Exec(ctx, query, formatISO8601(&now), t.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func (t *Team) InvitePlayer(
	ctx context.Context,
	tx *db.SafeWTX,
	playerID uint16,
) (*PlayerTeamInvite, error) {
	query := `
INSERT INTO player_team_invite (player_id, team_id, approved)
SELECT ?, ?, 
    CASE
        WHEN EXISTS (
            SELECT 1 FROM team_registration tr
            JOIN season s ON tr.season_id = s.id
            WHERE s.active = 1
            AND tr.team_id = ?
            AND tr.approved = 1
        ) THEN NULL
        ELSE 1
    END;`
	result, err := tx.Exec(ctx, query, playerID, t.ID, t.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Exec")
	}
	inviteID, err := result.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "result.LastInsertId")
	}
	invite, err := GetPlayerTeamInvite(ctx, tx, uint32(inviteID))
	if err != nil {
		return nil, errors.Wrap(err, "GetPlayerTeamInvite")
	}

	return invite, nil
}

func (t *Team) InvitedPlayers(
	ctx context.Context,
	tx db.SafeTX,
) (*[]PlayerTeamInvite, error) {
	query := `
SELECT pti.id, pti.player_id, p.name, pti.team_id, t.name, pti.status, pti.approved
FROM player_team_invite pti
JOIN player p ON pti.player_id = p.id
JOIN team t ON pti.team_id = t.id
WHERE pti.team_id = ? AND (
    (pti.status IS NULL AND pti.approved IS NULL) OR
    (pti.status IS NULL AND pti.approved = 1) OR
    (pti.status = 1 AND pti.approved IS NULL)
);`
	rows, err := tx.Query(ctx, query, t.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	invitedPlayers := []PlayerTeamInvite{}
	for rows.Next() {
		var playerinv PlayerTeamInvite
		var status sql.NullInt16
		var approved sql.NullInt16
		err = rows.Scan(
			&playerinv.ID,
			&playerinv.PlayerID,
			&playerinv.PlayerName,
			&playerinv.TeamID,
			&playerinv.TeamName,
			&status,
			&approved,
		)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		if status.Valid {
			statusInt := uint16(status.Int16)
			playerinv.Status = &statusInt
		}
		if approved.Valid {
			approvedInt := uint16(approved.Int16)
			playerinv.Approved = &approvedInt
		}
		invitedPlayers = append(invitedPlayers, playerinv)
	}
	return &invitedPlayers, nil
}

func (t *Team) RevokeInvite(
	ctx context.Context,
	tx *db.SafeWTX,
	playerID uint16,
) error {
	query := `
DELETE FROM player_team_invite WHERE team_id = ? AND player_id = ?
AND (approved IS NULL OR status IS NULL);`
	_, err := tx.Exec(ctx, query, t.ID, playerID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func (t *Team) RegistrationStatus(
	ctx context.Context,
	tx db.SafeTX,
) (*TeamRegistration, error) {
	query := `
SELECT tr.id, tr.team_id, tr.season_id, s.name, tr.preferred_league, tr.approved, tr.placed
FROM team_registration tr
JOIN season s ON tr.season_id = s.id
WHERE team_id = ? AND s.active = 1
AND (tr.approved IS NULL OR tr.approved = 1);
`
	row, err := tx.QueryRow(ctx, query, t.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var teamreg TeamRegistration
	var approved sql.NullInt16
	err = row.Scan(
		&teamreg.ID,
		&teamreg.TeamID,
		&teamreg.SeasonID,
		&teamreg.SeasonName,
		&teamreg.PreferredLeague,
		&approved,
		&teamreg.Placed,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	if approved.Valid {
		appr := uint16(approved.Int16)
		teamreg.Approved = &appr
	}
	return &teamreg, nil
}

func (t *Team) GetManager(ctx context.Context, tx db.SafeTX) (*Player, error) {
	query := `SELECT id, slap_id, name, discord_id FROM player WHERE id = ?;`
	row, err := tx.QueryRow(ctx, query, t.ManagerID)
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

func (t *Team) SetColor(ctx context.Context, tx *db.SafeWTX, hexStr string) error {
	matched, _ := regexp.MatchString("^[0-9a-fA-F]{6}$", hexStr)
	if !matched {
		return errors.New("Invalid hex string provided")
	}
	query := `UPDATE team SET color = ? WHERE id = ?;`
	_, err := tx.Exec(ctx, query, hexStr, t.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	hexInt, err := hexToInt(hexStr)
	if err != nil {
		return errors.Wrap(err, "hexToInt")
	}
	t.Color = hexInt
	return nil
}

func (t *Team) Register(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	preferredLeague string,
) (*TeamRegistration, error) {
	if preferredLeague != "Open" && preferredLeague != "IM" && preferredLeague != "Pro" {
		return nil, errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	query := `
INSERT INTO team_registration(team_id, season_id, preferred_league)
VALUES (?, ?, ?);
`
	res, err := tx.Exec(ctx, query, t.ID, seasonID, preferredLeague)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Exec")
	}
	trID, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "res.LastInsertId")
	}
	query = `
SELECT tr.id, tr.season_id, s.name
FROM team_registration tr
JOIN season s ON tr.season_id = s.id
WHERE tr.id = ?;
`
	row, err := tx.QueryRow(ctx, query, trID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var tr TeamRegistration
	err = row.Scan(&tr.ID, &tr.SeasonID, &tr.SeasonName)
	if err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}
	tr.TeamID = t.ID
	tr.TeamName = t.Name
	tr.PreferredLeague = preferredLeague
	tr.Placed = 0
	tr.PlacedLeagueName = ""
	return &tr, nil
}
