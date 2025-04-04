package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectSeason = &bot.Message{
	Label:       "Select Active Season",
	Purpose:     models.MsgSelectSeason,
	GetContents: selectSeasonComponents,
}

// Get the message contents for the select active season component
func selectSeasonComponents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select season components")
	seasons, err := models.GetSeasons(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetSeasons")
	}
	activeSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	options := []discordgo.SelectMenuOption{
		{
			Label:   "No active season",
			Value:   "NOACTIVESEASON",
			Default: activeSeason == nil,
		},
	}
	for _, season := range seasons {
		options = append(options, discordgo.SelectMenuOption{
			Label:   season.Name,
			Value:   season.ID,
			Default: season.Active,
		})
	}
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Current Season",
			Description: `
Select the season to be set as the active season.

**NOTE**
This will update all related messages to show data for the selected season.
(i.e. team rosters, fixtures).

**WARNING**
This will NOT change, add or remove Team/FA roles from players.
Changing off an active season before it is finished is a BAD IDEA and you shouldn't do it.`,
			Color: 0x00ff00, // Green color
		},
		Components: components.ActionRow(components.StringSelect(
			"season_select",
			"Select active season",
			options,
			1,
			1,
			false,
		)),
	}
	return contents, nil
}
