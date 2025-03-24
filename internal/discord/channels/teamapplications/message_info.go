package teamapplications

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

var infoMessage = &bot.Message{
	Label:       "Team Application Review Info",
	Purpose:     models.MsgTeamAppsInfo,
	GetContents: infoMessageContents,
}

func infoMessageContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Team Application Approvals!",
			Description: `
**This is where team applications will be sent!**

__Approving:__ 
Once an application is approved, it will lock their submitted roster and notify the manager.
The team will appear with their given roster in the Team Rosters channel.

__Placement:__
Once placed, the team will be entered into the selected league and the manager will be notified.
The Team Rosters channel will update to show the placement.`,
			Color: 0x00ff00, // Green color
		},
		Components: []discordgo.MessageComponent{},
	}
	return contents, nil
}
