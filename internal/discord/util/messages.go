package util

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Function for returning message contents for a complex message
type MessageContents func() (string, *discordgo.MessageEmbed, []discordgo.MessageComponent)

// Function that returns a MessageContents with the context and bot provided
type MessageContentsFunc func(ctx context.Context, b *Bot) (MessageContents, error)

// create a complex message. duplicate of the function in the messages package
// copied here to prevent broken import cycle
func createComplexMessage(
	contents MessageContents,
	channelID string,
	session *discordgo.Session,
) (string, error) {
	msg, embed, components := contents()
	message, err := session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content:    msg,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return "", errors.Wrap(err, "session.ChannelMessageSendComplex")
	}
	return message.ID, nil
}
