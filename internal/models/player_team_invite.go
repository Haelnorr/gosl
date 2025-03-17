package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Model of the player_team_invite table in the database
// Each row represents an invite sent to a player to join a team
type PlayerTeamInvite struct {
	ID         uint32  // unique ID
	PlayerID   uint16  // FK -> Player.ID
	PlayerName string  // from Player.Name
	TeamID     uint16  // FK -> Team.ID
	TeamName   string  // from Team.Name
	Status     *uint16 // nil for pending, 0 for rejected, 1 for accepted
	Approved   *uint16 // nil for pending, 0 for denied, 1 for approved
}

func GetPlayerTeamInvite(
	ctx context.Context,
	tx db.SafeTX,
	inviteID uint32,
) (*PlayerTeamInvite, error) {
	query := `
SELECT pti.id, pti.player_id, p.name, pti.team_id, t.name, pti.status, pti.approved 
FROM player_team_invite pti
JOIN player p ON pti.player_id = p.id
JOIN team t ON pti.team_id = t.id
WHERE pti.id = ?;`
	row, err := tx.QueryRow(ctx, query, inviteID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	var pti PlayerTeamInvite
	var status sql.NullInt16
	var approved sql.NullInt16
	err = row.Scan(
		&pti.ID,
		&pti.PlayerID,
		&pti.PlayerName,
		&pti.TeamID,
		&pti.TeamName,
		&status,
		&approved,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, errors.Wrap(err, "row.Scan")
	}
	if status.Valid {
		statusint := uint16(status.Int16)
		pti.Status = &statusint
	}
	if approved.Valid {
		approvedint := uint16(approved.Int16)
		pti.Approved = &approvedint
	}

	return &pti, nil
}

func (i *PlayerTeamInvite) Accept(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE player_team_invite SET status = 1 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, i.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	status := uint16(1)
	i.Status = &status
	return nil
}

func (i *PlayerTeamInvite) Approve(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE player_team_invite SET approved = 1 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, i.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(1)
	i.Status = &approved
	return nil
}

func (i *PlayerTeamInvite) Reject(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE player_team_invite SET status = 0 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, i.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	status := uint16(0)
	i.Status = &status
	return nil
}

func (i *PlayerTeamInvite) Deny(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE player_team_invite SET approved = 0 WHERE id = ?;`
	_, err := tx.Exec(ctx, query, i.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	approved := uint16(0)
	i.Status = &approved
	return nil
}
