package managerchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectSeason = &messages.ChannelMessage{
	Label:        "Select Active Season",
	Purpose:      messages.ManagerSelectSeason,
	Channel:      channels.PurposeManager,
	ContentsFunc: selectSeasonComponents,
}

// Get the message contents for the select active season component
func selectSeasonComponents(
	ctx context.Context,
	b *util.Bot,
) (util.MessageContents, error) {
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
	noActiveSeason := false
	if activeSeason.ID == "NOACTIVESEASON" {
		noActiveSeason = true
	}
	options := []discordgo.SelectMenuOption{
		{
			Label:   "No active season",
			Value:   "NOACTIVESEASON",
			Default: noActiveSeason,
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
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retreiving select season components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Current Season",
				Description: `
Select the season to be set as the active season.

**NOTE**
This will update all related messages to show data for the selected season.
(i.e. team rosters, fixtures).`,
				Color: 0x00ff00, // Green color
			},
			messages.StringSelect(
				"season_select",
				"Select active season",
				options,
				1,
				1,
			)
	}, nil
}

func handleSelectSeasonInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	season := i.MessageComponentData().Values[0]
	err := models.SetActiveSeason(ctx, tx, season)
	if err != nil {
		return errors.Wrap(err, "models.SetActiveSeason")
	}

	msg := "Active season set to: " + season
	b.Log().UserEvent(i.Member, msg)
	messages.ReplyEphemeral(msg, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating season select")
		err := messages.UpdateChannelMessage(ctx, b, selectSeason)
		if err != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select active season message after interaction")
		}
		b.Logger.Debug().Msg("Updating active season info")
		err = messages.UpdateChannelMessage(ctx, b, activeSeasonInfo)
		if err != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update active season message after interaction")
		}
		// TODO: update any other messages that display data from the active season
	}()
	return nil
}
