package bot

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	messageAdminSelectLogChannel   uint16 = 1
	messageAdminSelectAdminRoles   uint16 = 2
	messageAdminSelectManagerRoles uint16 = 3
	messageAdminTestButton         uint16 = 4
)

type MessageContents func() (string, *discordgo.MessageEmbed, []discordgo.MessageComponent)
type MessageContentsFunc func(ctx context.Context) (MessageContents, error)

func addMessagePurpose(
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

func removeMessagePurpose(
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

// Returns messageID, channelID, err
func getMessageForPurpose(
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

func (b *Bot) checkMessageExists(messageID, channelID string) bool {
	_, err := b.session.ChannelMessage(channelID, messageID)
	return err == nil
}

func (b *Bot) createStaticMessage(
	contents MessageContents,
	channelID string,
) (string, error) {
	msg, embed, components := contents()
	message, err := b.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    msg,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return "", errors.Wrap(err, "session.ChannelMessageSendComplex")
	}
	return message.ID, nil
}

func (b *Bot) editStaticMessage(
	contents MessageContents,
	messageID string,
	channelID string,
) error {
	msg, embed, components := contents()
	_, err := b.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Content:    &msg,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	if err != nil {
		return errors.Wrap(err, "session.ChannelMessageEditComplex")
	}
	return nil
}

func (b *Bot) updateChannelMessage(
	ctx context.Context,
	contentsFunc MessageContentsFunc,
	purpose uint16,
	channelID string,
) error {
	components, err := contentsFunc(ctx)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("contentsFunc (purpose %v)", purpose))
	}
	err = b.addOrEditChannelMessage(
		ctx, channelID, purpose, components)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("updateChannelMessage (purpose %v)", purpose))
	}
	return nil
}

func (b *Bot) addOrEditChannelMessage(
	ctx context.Context,
	defChannelID string,
	purpose uint16,
	contents MessageContents,
) error {
	b.logger.Debug().
		Uint16("purpose", purpose).
		Str("channel", defChannelID).
		Msg("Updating message")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	defer tx.Commit()
	b.logger.Debug().Uint16("purpose", purpose).Msg("Finding existing message")
	messageID, channelID, err := getMessageForPurpose(
		ctx, tx, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "getMessageForPurpose")
	}
	if messageID != "" && channelID != "" {
		if exists := b.checkMessageExists(messageID, channelID); exists {
			b.logger.Debug().Uint16("purpose", purpose).Msg("Message found, editing")
			err = b.editStaticMessage(contents, messageID, channelID)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "b.editStaticMessage")
			}
			return nil
		}
	}
	b.logger.Debug().Uint16("purpose", purpose).Msg("No message found, creating new message")
	messageID, err = b.createStaticMessage(contents, defChannelID)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "b.createStaticMessage")
	}
	b.logger.Debug().Uint16("purpose", purpose).Msg("Adding message to database")
	err = addMessagePurpose(ctx, tx, messageID, defChannelID, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "addMessagePurpose")
	}
	return nil
}
