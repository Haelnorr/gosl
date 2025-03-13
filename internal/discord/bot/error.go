package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Send an ephemeral error message to the user with details of the error
func (b *Bot) Error(
	pMsg string,
	sMsg string,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	embed := &discordgo.MessageEmbed{
		Color: 0xff1919,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    "Error",
			IconURL: "attachment://error.png",
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   pMsg,
				Value:  sMsg,
				Inline: false,
			},
		},
	}
	errIco, err := GetAsset("error.png", b.Files)
	if err != nil {
		return errors.Wrap(err, "getAsset")
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Files:  []*discordgo.File{errIco},
			Flags:  discordgo.MessageFlagsEphemeral,
		},
	})

}

// Helper function to send a Error response to the user advising them they
// do not have permission to perform the action
func (b *Bot) Forbidden(
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	msg := "You do not have permission for this action"
	return b.Error("Forbidden", msg, s, i)
}
