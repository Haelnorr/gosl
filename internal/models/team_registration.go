package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Model of the team_registration table in the database
// Each row represents a teams application to play in a season
type TeamRegistration struct {
	ID               uint16  // unique ID
	TeamID           uint16  // FK -> Team.ID
	TeamName         string  // from Team.Name
	ManagerID        string  // from Team.ManagerID
	SeasonID         string  // FK -> Season.ID
	SeasonName       string  // From Season.Name
	PreferredLeague  string  // League the team prefers to play in
	Approved         *uint16 // nil for pending, 0 for denied, 1 for accepted
	Placed           uint16  // 0 for not placed, League.ID for placed
	PlacedLeagueName string  // From League.Name
}

func GetTeamRegistration(
	ctx context.Context,
	tx db.SafeTX,
	appID uint16,
) (*TeamRegistration, error) {
	query := `
SELECT tr.id, t.id, t.name, p.discord_id, s.id, s.name, tr.preferred_league,
    tr.approved, tr.placed, l.division
FROM team_registration tr
JOIN team t ON tr.team_id = t.id
JOIN season s ON tr.season_id = s.id
JOIN player p ON t.manager_id = p.id
LEFT JOIN league l ON tr.placed = l.id
WHERE tr.id = ?;
`
	row, err := tx.QueryRow(ctx, query, appID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var tr TeamRegistration
	var approved sql.NullInt16
	var league sql.NullString
	err = row.Scan(
		&tr.ID,
		&tr.TeamID,
		&tr.TeamName,
		&tr.ManagerID,
		&tr.SeasonID,
		&tr.SeasonName,
		&tr.PreferredLeague,
		&approved,
		&tr.Placed,
		&league,
	)
	if err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}
	if approved.Valid {
		appr := uint16(approved.Int16)
		tr.Approved = &appr
	}
	if league.Valid {
		tr.PlacedLeagueName = league.String
	} else {
		tr.PlacedLeagueName = "Not yet placed"
	}
	return &tr, nil
}

func (tr *TeamRegistration) Approve(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE team_registration SET approved = 1 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, tr.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(1)
	tr.Approved = &approved
	return nil
}

func (tr *TeamRegistration) Reject(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE team_registration SET approved = 0 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, tr.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(0)
	tr.Approved = &approved
	return nil
}

func (tr *TeamRegistration) Place(
	ctx context.Context,
	tx *db.SafeWTX,
	leagueID uint16,
) error {
	var exists int
	query := `
SELECT EXISTS (
    SELECT 1 FROM team_league tl
    JOIN league l ON tl.league_id = l.id
    JOIN season s ON l.season_id = s.id
    WHERE tl.team_id = ?
    AND l.enabled = 1
);
`
	row, err := tx.QueryRow(ctx, query, tr.TeamID)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&exists)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	if exists == 1 {
		return errors.New("VE:Team already placed into a League for this season")
	}
	var enabled int
	query = `
SELECT EXISTS (
    SELECT 1 FROM league WHERE id = ? AND enabled = 1
);
`
	row, err = tx.QueryRow(ctx, query, leagueID)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&enabled)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	if enabled == 0 {
		return errors.New("VE:League is not enabled")
	}
	query = `INSERT INTO team_league(team_id, league_id) VALUES (?,?);`
	_, err = tx.Exec(ctx, query, tr.TeamID, leagueID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	query = `UPDATE team_registration SET placed = ? WHERE id = ?;`
	_, err = tx.Exec(ctx, query, leagueID, tr.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	query = `SELECT division FROM league WHERE id = ?;`
	row, err = tx.QueryRow(ctx, query, leagueID)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	var name string
	err = row.Scan(&name)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	tr.PlacedLeagueName = name
	tr.Placed = leagueID
	return nil
}
