package bot

import (
	"context"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

const (
	channelAdmin         uint16 = 1
	channelLog           uint16 = 2
	channelLeagueManager uint16 = 3
)

func addPurpose(ctx context.Context, tx *db.SafeTX, channelID string, purpose uint16) error {
	query := `INSERT INTO config_channels (channel_id, purpose) VALUES (?, ?) ON CONFLICT DO NOTHING;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

func removePurpose(ctx context.Context, tx *db.SafeTX, channelID string, purpose uint16) error {
	query := `DELETE FROM config_channels WHERE channel_id = ? AND purpose = ?;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

func getChannelsForPurpose(
	ctx context.Context,
	tx *db.SafeTX,
	purpose uint16,
) ([]string, error) {
	query := `SELECT channel_id FROM config_channels WHERE purpose = ?;`
	rows, err := tx.Query(ctx, query, purpose)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channelIDs []string
	for rows.Next() {
		var channelID string
		if err := rows.Scan(&channelID); err != nil {
			return nil, err
		}
		channelIDs = append(channelIDs, channelID)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows.Next")
	}

	return channelIDs, nil
}
