package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle an interaction with the select log channel component
func handleSelectLogChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	msgSelectLogChannel, err := b.GetMessage(models.ChannelAdmin, models.MsgSelectLogChannel)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	msgSelectLogChannel.StartUpdate(false)
	selectedChannel := i.MessageComponentData().Values[0]

	err = b.Channels[models.ChannelLog].UpdateTarget(ctx, tx, selectedChannel)
	if err != nil {
		return errors.Wrap(err, "channel.UpdateTarget (log channel)")
	}

	channel := i.MessageComponentData().Resolved.Channels[selectedChannel]
	msg := "Log channel updated to: " + channel.Name
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp("Updated log channel to "+channel.Name, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating log channel select")
		errch := make(chan error)
		go msgSelectLogChannel.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update message after interaction"
				b.DoubleError(msg, err)
			}
		}
	}()
	return nil
}
