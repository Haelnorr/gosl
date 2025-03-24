package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Model of the free_agent_registration table in the database
// Each row represents a request for a player to play as a free agent in a season
type FreeAgentRegistration struct {
	ID               uint32  // unique ID
	PlayerID         uint16  // FK -> Player.ID
	PlayerName       string  // from Player.Name
	SeasonID         string  // FK -> Season.ID
	SeasonName       string  // from Season.Name
	PreferredLeague  string  // League the player prefers to play in
	Approved         *uint16 // nil for pending, 0 for denied, 1 for accepted
	Placed           uint16  // has the player been placed into a league?
	PlacedLeagueName string  // from League.Division
}

// Returns true if already registered, false if not
func CheckPlayerFreeAgentRegistration(
	ctx context.Context,
	tx db.SafeTX,
	playerID uint16,
	seasonID string,
) (bool, error) {
	query := `
SELECT EXISTS (
    SELECT 1 FROM free_agent_registration
    WHERE player_id = ?
    AND season_id = ?
    AND (approved IS NULL OR approved = 1)
);`
	var exists int
	row, err := tx.QueryRow(ctx, query, playerID, seasonID)
	if err != nil {
		return false, errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "row.Scan")
	}
	if exists == 1 {
		return true, nil
	} else {
		return false, nil
	}
}

func GetFreeAgentRegistration(
	ctx context.Context,
	tx db.SafeTX,
	appID uint32,
) (*FreeAgentRegistration, error) {
	query := `
SELECT fa.id, fa.player_id, p.name, fa.season_id, s.name, fa.preferred_league,
    fa.approved, fa.placed, l.division
FROM free_agent_registration fa
JOIN player p ON fa.player_id = p.id
JOIN season s ON fa.season_id = s.id
LEFT JOIN league l ON fa.placed = l.id
WHERE fa.id = ?;
`
	row, err := tx.QueryRow(ctx, query, appID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var fa FreeAgentRegistration
	var approved sql.NullInt16
	var division sql.NullString
	err = row.Scan(&fa.ID, &fa.PlayerID, &fa.PlayerName, &fa.SeasonID,
		&fa.SeasonName, &fa.PreferredLeague, &approved, &fa.Placed, &division)
	if err != nil {
		return nil, errors.Wrap(err, "row.Scan")
	}
	if approved.Valid {
		appr := uint16(approved.Int16)
		fa.Approved = &appr
	}
	if division.Valid {
		fa.PlacedLeagueName = division.String
	} else {
		fa.PlacedLeagueName = "Not yet placed"
	}
	return &fa, nil
}

func (fa *FreeAgentRegistration) Approve(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE free_agent_registration SET approved = 1 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, fa.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(1)
	fa.Approved = &approved
	return nil
}

func (fa *FreeAgentRegistration) Reject(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE free_agent_registration SET approved = 0 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, fa.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(0)
	fa.Approved = &approved
	return nil
}

func (fa *FreeAgentRegistration) Place(
	ctx context.Context,
	tx *db.SafeWTX,
	leagueID uint16,
) error {
	var exists int
	query := `
SELECT EXISTS (
    SELECT 1 FROM free_agent fa
    JOIN league l ON fa.league_id = l.id
    JOIN season s ON l.season_id = s.id
    WHERE fa.player_id = ?
    AND l.enabled = 1
);
`
	row, err := tx.QueryRow(ctx, query, fa.PlayerID)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&exists)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	if exists == 1 {
		return errors.New("VE:Player already placed into a League for this season")
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
	query = `INSERT INTO free_agent(player_id, league_id) VALUES (?,?);`
	_, err = tx.Exec(ctx, query, fa.PlayerID, leagueID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	query = `UPDATE free_agent_registration SET placed = ? WHERE id = ?;`
	_, err = tx.Exec(ctx, query, leagueID, fa.ID)
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
	fa.PlacedLeagueName = name
	fa.Placed = leagueID
	return nil
}
