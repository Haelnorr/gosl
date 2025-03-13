package messages

import (
	"fmt"
	"gosl/internal/discord/util"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Create a complex message and return the messageID
func CreateComplexMessage(
	contents util.MessageContents,
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

// Edit a complex message
func EditComplexMessage(
	contents util.MessageContents,
	messageID string,
	channelID string,
	session *discordgo.Session,
) error {
	msg, embed, components := contents()
	starttime := time.Now()
	fmt.Println("Starting timer")
	_, err := session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         messageID,
		Channel:    channelID,
		Content:    &msg,
		Embeds:     &[]*discordgo.MessageEmbed{embed},
		Components: &components,
	})
	fmt.Printf("Time taken: %s\n", time.Since(starttime))
	if err != nil {
		return errors.Wrap(err, "session.ChannelMessageEditComplex")
	}
	return nil
}
