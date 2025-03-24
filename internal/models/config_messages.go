package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

const (
	// Admin channel messages
	MsgSelectLogChannel uint16 = 1 // select log channel message
	MsgSelectRoles      uint16 = 2 // select manager roles message
	MsgSelectChannels   uint16 = 3 // select registration channel message

	// Manager channel messages
	MsgSelectSeason uint16 = 11 // select season message
	MsgCreateSeason uint16 = 12 // create season message
	MsgActiveSeason uint16 = 13 // active season message

	// Registration channel messages
	MsgPlayerRegistration    uint16 = 21 // player registration message
	MsgTeamRegistration      uint16 = 22 // team registration message
	MsgFreeAgentRegistration uint16 = 23 // free agent registration message

	// Team applications channel messages
	MsgTeamAppsInfo uint16 = 31 // team applications channel information

	// Team Rosters channel messages
	MsgTeamRosters uint16 = 41 // team and free agent rosters for the current season

	// Transfer Approvals Channel messages
	MsgTransferApprovalsInfo uint16 = 51 // transfer approvals channel information

	// Free agent applications channel messages
	MsgFreeAgentAppsInfo uint16 = 61 // free agent applications channel information
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
