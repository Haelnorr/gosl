package directmessages

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

func confirmLeaveTeamComponents(team *models.Team, panelMsgID string) *bot.MessageContents {
	embed := &discordgo.MessageEmbed{
		Color: 0xff1919,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: fmt.Sprintf("Confirm leave %s (%s)?", team.Name, team.Abbreviation),
				Value: fmt.Sprintf(`
Are you sure you want to leave your current team %s?
`, team.Name),
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("confirm_leave_team_%v", panelMsgID),
					Label:    "Confirm Leave Team",
					Style:    discordgo.DangerButton,
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
