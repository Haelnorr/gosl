package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type DynamicMessage struct {
	ID        string
	ChannelID string
	Label     string
	b         *Bot
}

func NewDynamicMessage(label, channelID string, b *Bot) *DynamicMessage {
	return &DynamicMessage{
		Label:     label,
		ChannelID: channelID,
		b:         b,
	}
}

func reAddDynamicMessage(label, messageID, channelID string, b *Bot) *DynamicMessage {
	msg := &DynamicMessage{
		ID:        messageID,
		Label:     label,
		ChannelID: channelID,
		b:         b,
	}
	b.addDynamicMessage(msg)
	return msg
}

func (dy *DynamicMessage) Send(contents *MessageContents) error {
	msg := ""
	message, err := dy.b.Session.ChannelMessageSendComplex(dy.ChannelID, &discordgo.MessageSend{
		Content:    msg,
		Embeds:     []*discordgo.MessageEmbed{contents.Embed},
		Components: contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageSendComplex (%s)", dy.Label))
	}
	dy.ID = message.ID
	err = dy.b.addDynamicMessage(dy)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("bot.AddDynamicMessage (%s)", dy.Label))
	}
	return nil
}

func (dy *DynamicMessage) Update(contents *MessageContents) error {
	if dy.ID == "" {
		return errors.New("Message has not been sent yet. Use Send() first")
	}
	msg := ""
	_, err := dy.b.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         dy.ID,
		Channel:    dy.ChannelID,
		Content:    &msg,
		Embeds:     &[]*discordgo.MessageEmbed{contents.Embed},
		Components: &contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageEditComplex (%s)", dy.Label))
	}
	return nil
}

func (dy *DynamicMessage) Expire(contents *MessageContents) error {
	msg := "*Message is locked*"
	disableComponents(&contents.Components)
	_, err := dy.b.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         dy.ID,
		Channel:    dy.ChannelID,
		Content:    &msg,
		Embeds:     &[]*discordgo.MessageEmbed{contents.Embed},
		Components: &contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageEditComplex (%s)", dy.Label))
	}
	err = dy.b.removeDynamicMessage(dy.ID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("bot.RemoveDynamicMessage (%s)", dy.Label))
	}
	return nil
}

func (dy *DynamicMessage) Delete(
	embed *discordgo.MessageEmbed,
	comps []discordgo.MessageComponent,
) error {
	err := dy.b.Session.ChannelMessageDelete(dy.ChannelID, dy.ID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageDelete (%s)", dy.Label))
	}
	err = dy.b.removeDynamicMessage(dy.ID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("bot.RemoveDynamicMessage (%s)", dy.Label))
	}
	return nil
}
