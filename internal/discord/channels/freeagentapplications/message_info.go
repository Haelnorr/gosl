package freeagentapplications

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

var infoMessage = &bot.Message{
	Label:       "Free Agent Application Review Info",
	Purpose:     models.MsgFreeAgentAppsInfo,
	GetContents: infoMessageContents,
}

func infoMessageContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Free Agent Approvals!",
			Description: `
**This is where free agent applications will be sent!**

__Approving:__ 
Once an application is approved, it will notify the player.
The player will appear under Free Agents in the Team Rosters channel.

__Placement:__
Once placed, the player will be entered into the selected league and be notified.
The Team Rosters channel will update to show the placement.`,
			Color: 0x00ff00, // Green color
		},
		Components: []discordgo.MessageComponent{},
	}
	return contents, nil
}
