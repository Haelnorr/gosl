package registrationchannel

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/pkg/steamapi"

	"github.com/bwmarrin/discordgo"
)

func confirmSlapIDContents(steamuser *steamapi.User, slapid uint32) *bot.MessageContents {
	embed := &discordgo.MessageEmbed{
		Color: 0xeb7d34,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    steamuser.PersonaName,
			IconURL: steamuser.Avatar,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Steam User Found",
				Value:  fmt.Sprintf("__SlapID:__ %v", slapid),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("confirm_slapid_%v", slapid),
					Label:    "Confirm",
					Style:    discordgo.SuccessButton,
				},
			},
		},
	}
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents
}
