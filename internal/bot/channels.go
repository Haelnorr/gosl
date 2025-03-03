package bot

import (
	"context"
	"database/sql"
	"gosl/pkg/db"
	"time"

	"github.com/pkg/errors"
)

const (
	channelAdmin         uint16 = 1
	channelLog           uint16 = 2
	channelLeagueManager uint16 = 3
)

func addChannelPurpose(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
	query := `INSERT INTO config_channels (channel_id, purpose) VALUES (?, ?) ON CONFLICT DO NOTHING;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

func setChannelPurpose(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
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

func removeChannelPurpose(ctx context.Context, tx *db.SafeWTX, channelID string, purpose uint16) error {
	query := `DELETE FROM config_channels WHERE channel_id = ? AND purpose = ?;`
	_, err := tx.Exec(ctx, query, channelID, purpose)
	return err
}

func queryChannelForPurpose(
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
func queryChannelsForPurpose(
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

func (b *Bot) checkChannelExists(channelID string) bool {
	if channelID == "" {
		return false
	}
	_, err := b.session.Channel(channelID)
	return err == nil
}

func (b *Bot) getChannel(ctx context.Context, purpose uint16) (string, error) {
	b.logger.Debug().Uint16("purpose", purpose).Msg("Getting channel ID")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	channelID, err := queryChannelForPurpose(ctx, tx, purpose)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	tx.Commit()
	return channelID, nil
}
