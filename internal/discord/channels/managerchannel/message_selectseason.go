package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"time"

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
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select season components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select season components")
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
	tx.Commit()
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Current Season",
			Description: `
Select the season to be set as the active season.

**NOTE**
This will update all related messages to show data for the selected season.
(i.e. team rosters, fixtures).`,
			Color: 0x00ff00, // Green color
		},
		Components: components.StringSelect(
			"season_select",
			"Select active season",
			options,
			1,
			1,
			false,
		),
	}
	return contents, nil
}
