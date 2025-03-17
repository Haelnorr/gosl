package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

var createSeason = &bot.Message{
	Label:       "Create Season",
	Purpose:     models.MsgCreateSeason,
	GetContents: createSeasonComponents,
}

// Get the message contents for the create season component
func createSeasonComponents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: "create_season_button",
					Label:    "Create Season",
				},
			},
		},
	}
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Create Season",
			Description: `
            Create a new season.
            Season ID and Name must be unique.`,
			Color: 0x00ff00, // Green color
		},
		Components: components,
	}
	return contents, nil
}
