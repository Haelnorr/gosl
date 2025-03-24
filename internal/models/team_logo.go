package models

import (
	"context"
	"gosl/pkg/db"
	"time"

	"github.com/pkg/errors"
)

func UploadTeamLogo(
	ctx context.Context,
	tx *db.SafeWTX,
	logoURL string,
	channelID string,
	messageID string,
	teamID uint16,
) error {
	query := `
INSERT INTO team_logo(url, team_id, message_id, channel_id, uploaded)
VALUES (?, ?, ?, ?, ?);
`
	now := time.Now().Unix()
	_, err := tx.Exec(ctx, query, logoURL, teamID, messageID, channelID, now)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}
