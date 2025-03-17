package directmessages

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

func disbandTeamComponents(team *models.Team, messageID string) *bot.MessageContents {
	embed := &discordgo.MessageEmbed{
		Title: "Disband Team",
		Description: fmt.Sprintf(`
Are you sure you want to disband %s?
This will remove all players including yourself.
You **will** be able to rejoin this team in the future if you wish.`, team.Name),
		Color: 0xff1919,
	}
	comps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: fmt.Sprintf("disband_team_confirm_%s", messageID),
					Label:    "Confirm Disband Team",
					Style:    discordgo.DangerButton,
				},
			},
		},
	}
	return &bot.MessageContents{
		Embed:      embed,
		Components: comps,
	}
}
