package adminchannel

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

var selectRegistrationChannel = &bot.Message{
	Label:       "Select Registration Channel",
	Purpose:     models.MsgSelectRegistrationChannel,
	GetContents: selectRegistrationChannelContents,
}

// Get the message contents for the select log channel component
func selectRegistrationChannelContents(
	ctx context.Context,
	b *bot.Bot,
) (bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select registration channel components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select log channel components")
	regChannelID, err := models.GetChannel(ctx, tx, models.ChannelRegistration)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	tx.Commit()
	var defaultValues []discordgo.SelectMenuDefaultValue
	if regChannelID != "" {
		defaultValues = append(defaultValues, discordgo.SelectMenuDefaultValue{
			ID:   regChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retrieving select registration channel components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Registration channel",
				Description: `Select the channel to display player/team registration`,
				Color:       0x00ff00, // Green color
			},
			components.ChannelSelect(
				"registration_channel_select",
				"Select the channel for registration",
				defaultValues,
				1,
				1,
				[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
			)
	}, nil
}

// Handle an interaction with the select log channel component
func handleSelectRegistrationChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	selectRegistrationChannel := b.Channels[models.ChannelAdmin].Messages[models.MsgSelectRegistrationChannel]
	selectRegistrationChannel.StartUpdate(false)
	selectedChannel := i.MessageComponentData().Values[0]
	err := models.SetChannel(ctx, tx, selectedChannel, models.ChannelRegistration)
	if err != nil {
		return errors.Wrap(err, "setChannelPurpose (registration channel)")
	}
	b.Channels[models.ChannelRegistration].UpdateTarget(ctx, tx, selectedChannel)
	channel := i.MessageComponentData().Resolved.Channels[selectedChannel]
	msg := "Registration channel updated to: " + channel.Name
	b.Log().UserEvent(i.Member, msg)
	b.Reply("Updated registration channel to "+channel.Name, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating registration channel select")
		errch := make(chan error)
		selectRegistrationChannel.Update(ctx, errch)
		if <-errch != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select registration channel message after interaction")
		}
		close(errch)
	}()
	return nil
}
