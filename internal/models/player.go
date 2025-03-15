package models

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/pkg/db"
	"strings"

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
