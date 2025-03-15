package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

const (
	// Admin channel messages
	MsgSelectLogChannel          uint16 = 1 // select log channel component
	MsgSelectAdminRoles          uint16 = 2 // select admin roles component
	MsgSelectManagerRoles        uint16 = 3 // select manager roles component
	MsgSelectRegistrationChannel uint16 = 8 // select registration channel component

	// Manager channel messages
	MsgSelectSeason uint16 = 4 // select season component
	MsgCreateSeason uint16 = 5 // create season component
	MsgActiveSeason uint16 = 6 // active season component

	// Registration channel messages
	MsgPlayerRegistration uint16 = 7 // player registration component
)

// Set the provided message as the message used for the provided purpose
// Only one message can be used for a given purpose at a time
// Setting a new message will overwrite the previous one
func SetMessage(
	ctx context.Context,
	tx *db.SafeWTX,
	messageID string,
	channelID string,
	purpose uint16,
) error {
	query := `
INSERT INTO config_messages (message_id, channel_id, purpose) 
VALUES (?, ?, ?) 
ON CONFLICT(purpose) DO UPDATE
SET message_id = excluded.message_id,
    channel_id = excluded.channel_id;
`
	_, err := tx.Exec(ctx, query, messageID, channelID, purpose)
	return err
}

// Remove the provided message from the database
func RemoveMessage(
	ctx context.Context,
	tx *db.SafeWTX,
	messageID string,
	channelID string,
	purpose uint16,
) error {
	query := `
DELETE FROM config_messages WHERE message_id = ? AND channel_id = ? AND purpose = ?;
`
	_, err := tx.Exec(ctx, query, messageID, channelID, purpose)
	return err
}

// Get the message that has been set for the provided purpose
// Returns messageID, channelID, err
func GetMessageForPurpose(
	ctx context.Context,
	tx db.SafeTX,
	purpose uint16,
) (string, string, error) {
	query := `
SELECT message_id, channel_id FROM config_messages WHERE purpose = ?;
`
	var messageID string
	var channelID string
	row, err := tx.QueryRow(ctx, query, purpose)
	if err != nil {
		return "", "", errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&messageID, &channelID)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", errors.Wrap(err, "row.Scan")
	}

	return messageID, channelID, nil
}
