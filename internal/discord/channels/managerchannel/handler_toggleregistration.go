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
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		errch := make(chan error)
		go activeSeasonInfo.Update(ctx, errch)
		go teamRegistration.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to message after interaction"
				b.Logger.Warn().Err(err).Msg(msg)
				b.Log().Error(msg, err)
			}
		}
	}()
	return nil
}
