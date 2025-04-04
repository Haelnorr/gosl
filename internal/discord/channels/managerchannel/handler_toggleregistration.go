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

func handleToggleRegistrationInteraction(
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
	teamRegistration, err := b.GetMessage(models.ChannelRegistration, models.MsgTeamRegistration)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	if !activeSeasonInfo.StartUpdate(false) || !teamRegistration.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	b.Logger.Debug().Msg("Getting active season")
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	b.Logger.Debug().Msg("Toggling active season registration")
	err = season.ToggleRegistration(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "season.ToggleRegistration")
	}

	msg := "Registration status for %s set to %s"
	msg = fmt.Sprintf(msg, season.Name, season.RegistrationStatusString())
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	b.Logger.Debug().Msg("Updating active season message")
	errch := make(chan error)
	go activeSeasonInfo.Update(ctx, tx, errch)
	go teamRegistration.Update(ctx, tx, errch)
	hadErr := false
	for err := range errch {
		if err != nil {
			hadErr = true
			msg := "Failed to message after interaction"
			b.DoubleError(msg, err)
		}
	}
	if hadErr {
		return errors.New("Failed to update message(s) after interaction")
	}
	return nil
}
