package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

const (
	ChannelAdmin                uint16 = 1 // Channel used for admin panel
	ChannelLog                  uint16 = 2 // Channel used for logging
	ChannelManager              uint16 = 3 // Channel used for league manager panel
	ChannelRegistration         uint16 = 4 // Channel used for player and team registrations
	ChannelRegistrationApproval uint16 = 5 // Channel used for approving team and free agent registrations
)

// Add a channel to the database with the provided purpose
func AddChannel(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
	query := `INSERT INTO config_channels (channel_id, purpose) VALUES (?, ?) ON CONFLICT DO NOTHING;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

// Set a channel in the database as the only channel with the provided purpose
func SetChannel(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
	var count int
	query := `SELECT COUNT(*) FROM config_channels WHERE purpose = ?;`
	row, err := tx.QueryRow(ctx, query, purpose)
	if err != nil {
		return errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&count)
	if err != nil {
		return errors.Wrap(err, "row.Scan")
	}
	switch {
	case count == 1:
		query = `UPDATE config_channels SET channel_id = ? WHERE purpose = ?;`
		_, err = tx.Exec(ctx, query, channelID, purpose)
	case count == 0:
		query = `INSERT INTO config_channels (channel_id, purpose) VALUES (?,?);`
		_, err = tx.Exec(ctx, query, channelID, purpose)
	default:
		return errors.Errorf("Invalid row count for purpose %v. Expecting 0 or 1, got %v", purpose, count)
	}
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

// Remove a channel from the database with the provided purpose
func RemoveChannel(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
	query := `DELETE FROM config_channels WHERE channel_id = ? AND purpose = ?;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

// Get a single channel that has the purpose provided set in the database
func GetChannel(
	ctx context.Context,
	tx db.SafeTX,
	purpose uint16,
) (string, error) {
	query := `SELECT channel_id FROM config_channels WHERE purpose = ? LIMIT 1;`
	row, err := tx.QueryRow(ctx, query, purpose)
	if err != nil {
		return "", err
	}
	var channelID string
	err = row.Scan(&channelID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", errors.Wrap(err, "row.Scan")
	}
	return channelID, nil
}

// Get all the channels from the database with the provided purpose
func GetChannels(
	ctx context.Context,
	tx db.SafeTX,
	purpose uint16,
) ([]string, error) {
	query := `SELECT channel_id FROM config_channels WHERE purpose = ?;`
	rows, err := tx.Query(ctx, query, purpose)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()

	var channelIDs []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		channelIDs = append(channelIDs, channelID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Next")
	}

	return channelIDs, nil
}
