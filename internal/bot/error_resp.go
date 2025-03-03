package bot

import (
	"io/fs"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func errorResponse(
	pMsg string,
	sMsg *string,
	files *fs.FS,
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
				Value:  "\u200b",
				Inline: false,
			},
		},
	}

	if sMsg != nil {
		embed.Fields[0].Value = *sMsg
	}
	errIco, err := getAsset("error.png", files)
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
