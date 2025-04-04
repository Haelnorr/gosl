package adminchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle an interaction with the select log channel component
func handleSelectChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	purpose uint16,
) error {
	b.Acknowledge(i, ack)
	msgSelectChannels, err := b.GetMessage(models.ChannelAdmin,
		models.MsgSelectChannels)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	if !msgSelectChannels.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	selectedChannel := i.MessageComponentData().Values[0]
	err = models.SetChannel(ctx, tx, selectedChannel, purpose)
	if err != nil {
		return errors.Wrap(err, "models.SetChannel")
	}
	channel, err := b.GetChannel(purpose)
	if err != nil {
		return errors.Wrap(err, "b.GetChannel")
	}
	err = channel.UpdateTarget(ctx, tx, selectedChannel)
	if err != nil {
		return errors.Wrap(err, "channel.UpdateTarget")
	}
	channelDiscord := i.MessageComponentData().Resolved.Channels[selectedChannel]
	msg := fmt.Sprintf("%s updated to: %s", channel.Label, channelDiscord.Name)
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	b.Logger.Debug().Msg("Updating channel select")
	errch := make(chan error)
	go msgSelectChannels.Update(ctx, tx, errch)
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
