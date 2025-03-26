package registrationchannel

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

func reregisterTeamComponents(team *models.Team) *bot.MessageContents {
	embed := &discordgo.MessageEmbed{
		Color: team.Color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("%s (%s)", team.Name, team.Abbreviation),
			IconURL: team.Logo,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: fmt.Sprintf("Currently manager of %s (%s)", team.Name, team.Abbreviation),
				Value: `
Do you want to re-register this team, or disband the team?

Re-registering will retain all current players and start the registration process.

Disbanding will remove all players (including you) and allow you to pick from any teams you have previously been a manager of.
`,
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("reregister_team_%v", team.ID),
					Label:    "Re-Register current team",
					Style:    discordgo.SuccessButton,
				},
				&discordgo.Button{
					CustomID: fmt.Sprintf("disband_team_%v", team.ID),
					Label:    "Disband current team",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
	return &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
}
