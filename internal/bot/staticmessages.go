package bot

import (
	"context"
	"database/sql"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	messageAdminSelectLogChannel   uint16 = 1
	messageAdminSelectAdminRoles   uint16 = 2
	messageAdminSelectManagerRoles uint16 = 3
)

type MessageContents func() (string, *discordgo.MessageEmbed, []discordgo.MessageComponent)

func addMessagePurpose(
	ctx context.Context,
	tx *db.SafeTX,
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
	tx *db.SafeTX,
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
	tx *db.SafeTX,
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
