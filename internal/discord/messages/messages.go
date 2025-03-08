package messages

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	AdminSelectLogChannel   uint16 = 1 // Admin channel: select log channel component
	AdminSelectAdminRoles   uint16 = 2 // Admin channel: select admin roles component
	AdminSelectManagerRoles uint16 = 3 // Admin channel: select manager roles component
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

// Check if a message exists with the discord API
func CheckMessageExists(messageID, channelID string, s *discordgo.Session) bool {
	_, err := s.ChannelMessage(channelID, messageID)
	return err == nil
}

// Update the provided message using the provided contents function
func UpdateChannelMessage(
	ctx context.Context,
	b *util.Bot,
	contentsFunc util.MessageContentsFunc,
	purpose uint16,
	channelID string,
) error {
	components, err := contentsFunc(ctx, b)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("contentsFunc (purpose %v)", purpose))
	}
	err = AddOrEditChannelMessage(
		ctx, b, channelID, purpose, components)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("AddOrEditChannelMessage (purpose %v)", purpose))
	}
	return nil
}

// Edit the message for the provided purpose; if it doesn't exist, create a new one
func AddOrEditChannelMessage(
	ctx context.Context,
	b *util.Bot,
	defChannelID string,
	purpose uint16,
	contents util.MessageContents,
) error {
	b.Logger.Debug().
		Uint16("purpose", purpose).
		Str("channel", defChannelID).
		Msg("Updating message")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	defer tx.Commit()
	b.Logger.Debug().Uint16("purpose", purpose).Msg("Finding existing message")
	messageID, channelID, err := GetMessageForPurpose(
		ctx, tx, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "getMessageForPurpose")
	}
	if messageID != "" && channelID != "" {
		if exists := CheckMessageExists(messageID, channelID, b.Session); exists {
			b.Logger.Debug().Uint16("purpose", purpose).Msg("Message found, editing")
			err = EditComplexMessage(contents, messageID, channelID, b.Session)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "b.editStaticMessage")
			}
			return nil
		}
	}
	b.Logger.Debug().Uint16("purpose", purpose).Msg("No message found, creating new message")
	messageID, err = CreateComplexMessage(contents, defChannelID, b.Session)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "b.createStaticMessage")
	}
	b.Logger.Debug().Uint16("purpose", purpose).Msg("Adding message to database")
	err = SetMessage(ctx, tx, messageID, defChannelID, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "addMessagePurpose")
	}
	return nil
}
