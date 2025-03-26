package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleSelectSeasonInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	msgSelectSeason, err := b.GetMessage(models.ChannelManager, models.MsgSelectSeason)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	msgActiveSeason, err := b.GetMessage(models.ChannelManager, models.MsgActiveSeason)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	teamRegistration, err := b.GetMessage(models.ChannelRegistration, models.MsgTeamRegistration)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	freeAgentRegistration, err := b.GetMessage(models.ChannelRegistration, models.MsgFreeAgentRegistration)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	teamRosters, err := b.GetMessage(models.ChannelTeamRosters, models.MsgTeamRosters)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	teamRegistration.StartUpdate(true)
	freeAgentRegistration.StartUpdate(true)
	teamRosters.StartUpdate(true)
	if !msgSelectSeason.StartUpdate(false) || !msgActiveSeason.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	season := i.MessageComponentData().Values[0]
	err = models.SetActiveSeason(ctx, tx, season)
	if err != nil {
		return errors.Wrap(err, "models.SetActiveSeason")
	}

	msg := "Active season set to: " + season
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		// NOTE: update any other messages that display data from the active season
		errch := make(chan error)
		go msgSelectSeason.Update(ctx, errch)
		go msgActiveSeason.Update(ctx, errch)
		go teamRegistration.Update(ctx, errch)
		go freeAgentRegistration.Update(ctx, errch)
		go teamRosters.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update message after interaction"
				b.Logger.Warn().Err(err).Msg(msg)
				b.Log().Error(msg, err)
			}
		}
	}()
	return nil
}
