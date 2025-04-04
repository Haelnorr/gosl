package managerchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleSelectLeaguesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	activeSeasonInfo, err := b.GetMessage(models.ChannelManager, models.MsgActiveSeason)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	if !activeSeasonInfo.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	b.Logger.Debug().Msg("Getting active season")
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	leagues := i.MessageComponentData().Values
	err = models.SetLeagues(ctx, tx, season.ID, leagues)
	if err != nil {
		return errors.Wrap(err, "models.SetLeagues")
	}

	msg := "Leagues updated for %s:\n"
	for _, league := range leagues {
		msg = msg + " - " + league + "\n"
	}
	msg = fmt.Sprintf(msg, season.Name)
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	b.Logger.Debug().Msg("Updating active season message")
	errch := make(chan error)
	go activeSeasonInfo.Update(ctx, tx, errch)
	hadErr := false
	for err := range errch {
		if err != nil {
			hadErr = true
			msg := "Failed to update message after interaction"
			b.DoubleError(msg, err)
		}
	}
	if hadErr {
		return errors.New("Failed to update message(s) after interaction")
	}
	return nil
}

func getLeagueOptions(
	leagues *[]models.League,
) ([]discordgo.SelectMenuOption, error) {
	allLeagues := map[string]*discordgo.SelectMenuOption{

		"Open": {
			Label: "Open",
			Value: "Open",
		},
		"IM": {
			Label: "Intermediate",
			Value: "IM",
		},
		"Pro": {
			Label: "Pro",
			Value: "Pro",
		},
	}
	options := []discordgo.SelectMenuOption{}
	for _, league := range *leagues {
		allLeagues[league.Division].Default = true
	}
	for _, option := range allLeagues {
		options = append(options, *option)
	}
	return options, nil
}
