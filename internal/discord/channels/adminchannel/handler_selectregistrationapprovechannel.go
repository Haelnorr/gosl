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
func handleSelectRegistrationApprovalChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	msgSelectChannels, err := b.GetMessage(models.ChannelAdmin,
		models.MsgSelectChannels)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	msgSelectChannels.StartUpdate(false)
	selectedChannel := i.MessageComponentData().Values[0]
	err = models.SetChannel(ctx, tx, selectedChannel, models.ChannelRegistrationApproval)
	if err != nil {
		return errors.Wrap(err, "setChannelPurpose (registration approval channel)")
	}
	b.Channels[models.ChannelRegistrationApproval].UpdateTarget(ctx, tx, selectedChannel)
	channel := i.MessageComponentData().Resolved.Channels[selectedChannel]
	msg := "Registration approvals channel updated to: " + channel.Name
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp("Updated registration approval channel to "+channel.Name, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating registration channel select")
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
