package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
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
) (bot.MessageContents, error) {
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
			components.StringSelect(
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
	b *bot.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	msgSelectSeason := b.Channels[models.ChannelManager].Messages[models.MsgSelectSeason]
	msgActiveSeason := b.Channels[models.ChannelManager].Messages[models.MsgActiveSeason]
	if !msgSelectSeason.StartUpdate(false) || !msgActiveSeason.StartUpdate(false) {
		b.Error("Slow down!", "An update is in progress, please try again", s, i)
		return nil
	}
	season := i.MessageComponentData().Values[0]
	err := models.SetActiveSeason(ctx, tx, season)
	if err != nil {
		return errors.Wrap(err, "models.SetActiveSeason")
	}

	msg := "Active season set to: " + season
	b.Log().UserEvent(i.Member, msg)
	bot.ReplyEphemeral(msg, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		// NOTE: update any other messages that display data from the active season
		errmsg := "Failed to update message after interaction"
		errch := make(chan error)
		go msgSelectSeason.Update(ctx, errch)
		go msgActiveSeason.Update(ctx, errch)

		for err := range errch {
			if err != nil {
				b.Logger.Warn().Err(err).Msg(errmsg)
			}
		}
	}()
	return nil
}
