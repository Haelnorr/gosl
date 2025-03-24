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
	msgSelectChannels.StartUpdate(false)
	selectedChannel := i.MessageComponentData().Values[0]
	err = models.SetChannel(ctx, tx, selectedChannel, purpose)
	if err != nil {
		return errors.Wrap(err, "models.SetChannel")
	}
	channel := b.Channels[purpose]
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
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating channel select")
		errch := make(chan error)
		go msgSelectChannels.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update message after interaction"
				b.DoubleError(msg, err)
			}
		}
	}()
	return nil
}
