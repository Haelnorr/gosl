package bot

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type DirectMessage struct {
	ID                string
	Label             string
	UserID            string        // Discord ID of the user the send the DM to
	Expiry            time.Duration // How long until message expires
	ExpiresAt         time.Time     // Timestamp of when to 'expire' the message
	deleteAfterExpiry bool
	b                 *Bot
	c                 *discordgo.Channel
}

// If 0 is provided as expiry time, will never expire
func NewDirectMessage(
	label string,
	userID string,
	expiry time.Duration,
	deleteAfterExpiry bool,
	b *Bot,
) *DirectMessage {
	return &DirectMessage{
		Label:             label,
		Expiry:            expiry,
		UserID:            userID,
		deleteAfterExpiry: deleteAfterExpiry,
		b:                 b,
	}
}

func reAddDirectMessage(
	label string,
	messageID string,
	userID string,
	channel *discordgo.Channel,
	expiry time.Duration,
	deleteAfterExpiry bool,
	b *Bot,
) (*DirectMessage, error) {
	msg := &DirectMessage{
		ID:                messageID,
		Label:             label,
		Expiry:            expiry,
		UserID:            userID,
		deleteAfterExpiry: deleteAfterExpiry,
		b:                 b,
		c:                 channel,
	}
	err := b.addDirectMessage(msg)
	if err != nil {
		return nil, errors.Wrap(
			err, fmt.Sprintf("bot.addDirectMessage (%s, %s)", label, userID))
	}
	return msg, nil
}
func (dm *DirectMessage) Send(contents *MessageContents) error {
	msg := ""
	if dm.Expiry != 0 {
		action := "lock"
		if dm.deleteAfterExpiry {
			action = "delete"
		}
		expiresAt := time.Now().Add(dm.Expiry)
		dm.ExpiresAt = expiresAt
		msg = fmt.Sprintf("*Message will %s %s*", action, DiscordUntil(&expiresAt))
	}
	channel, err := dm.b.Session.UserChannelCreate(dm.UserID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.UserChannelCreate (%s, %s)", dm.Label, dm.UserID))
	}
	dm.c = channel
	message, err := dm.b.Session.ChannelMessageSendComplex(dm.c.ID, &discordgo.MessageSend{
		Content:    msg,
		Embeds:     []*discordgo.MessageEmbed{contents.Embed},
		Components: contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageSendComplex (%s, %s)", dm.Label, dm.UserID))
	}
	dm.ID = message.ID
	err = dm.b.addDirectMessage(dm)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("bot.addDirectMessage (%s, %s)", dm.Label, dm.UserID))
	}
	if dm.Expiry != 0 {
		go dm.waitForExpiry(contents)
	}
	return nil
}

func (dm *DirectMessage) waitForExpiry(contents *MessageContents) {
	time.Sleep(time.Until(dm.ExpiresAt))
	if time.Now().Before(dm.ExpiresAt) {
		go dm.waitForExpiry(contents)
		return
	}
	err := dm.Expire(contents)
	if err != nil {
		dm.b.Logger.Warn().Err(err).
			Str("msg", dm.Label).
			Str("user", dm.UserID).
			Msg("Failed to expire direct message")
	}
}

func (dm *DirectMessage) Update(contents *MessageContents) error {
	if dm.ID == "" {
		return errors.New("DM has not been sent yet. Use DirectMessage.Send() first")
	}
	msg := ""
	if dm.Expiry != 0 {
		action := "lock"
		if dm.deleteAfterExpiry {
			action = "delete"
		}
		expiresAt := time.Now().Add(dm.Expiry)
		dm.ExpiresAt = expiresAt
		msg = fmt.Sprintf("*Message will %s %s*", action, DiscordUntil(&expiresAt))
	}
	_, err := dm.b.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         dm.ID,
		Channel:    dm.c.ID,
		Content:    &msg,
		Embeds:     &[]*discordgo.MessageEmbed{contents.Embed},
		Components: &contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageEditComplex (%s, %s)", dm.Label, dm.UserID))
	}
	return nil
}

func (dm *DirectMessage) Expire(contents *MessageContents) error {
	if dm.deleteAfterExpiry {
		err := dm.b.Session.ChannelMessageDelete(dm.c.ID, dm.ID)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageDelete (%s, %s)", dm.Label, dm.UserID))
		}
	} else {
		msg := "*Message is locked*"
		disableComponents(&contents.Components)
		_, err := dm.b.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         dm.ID,
			Channel:    dm.c.ID,
			Content:    &msg,
			Embeds:     &[]*discordgo.MessageEmbed{contents.Embed},
			Components: &contents.Components,
		})
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageEditComplex (%s, %s)", dm.Label, dm.UserID))
		}
	}
	err := dm.b.removeDirectMessage(dm.ID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("bot.RemoveDirectMessage (%s, %s)", dm.Label, dm.UserID))
	}
	return nil
}

func disableComponents(comps *[]discordgo.MessageComponent) {
	for _, comp := range *comps {
		switch v := comp.(type) {
		case *discordgo.ActionsRow:
			disableComponents(&v.Components)
		case *discordgo.Button:
			v.Disabled = true
		case *discordgo.SelectMenu:
			v.Disabled = true
		case *discordgo.TextInput:
		}
	}
}
