package adminchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Get the message contents for the select log channel component
func selectLogChannelContents(
	ctx context.Context,
	b *util.Bot,
) (util.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select log channel components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select log channel components")
	logChannelID, err := channels.GetChannel(ctx, tx, channels.PurposeLog)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelForPurpose")
	}
	tx.Commit()
	var defaultValues []discordgo.SelectMenuDefaultValue
	if channels.CheckExists(logChannelID, b.Session) {
		defaultValues = append(defaultValues, discordgo.SelectMenuDefaultValue{
			ID:   logChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retrieving select log channel components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Log output channel",
				Description: `Select the channel to output bot logs to`,
				Color:       0x00ff00, // Green color
			},
			messages.ChannelSelect(
				"log_channel_select",
				"Select the channel for log output",
				defaultValues,
				1,
				1,
				[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
			)
	}, nil
}

// Handle an interaction with the select log channel component
func handleSelectLogChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	selectedChannel := i.MessageComponentData().Values[0]
	err := channels.SetPurpose(ctx, tx, selectedChannel, channels.PurposeLog)
	if err != nil {
		return errors.Wrap(err, "setChannelPurpose (log channel)")
	}
	b.Logchannel = selectedChannel
	channel := i.MessageComponentData().Resolved.Channels[selectedChannel]
	msg := "Log channel updated to: " + channel.Name
	b.Log().UserEvent(i.Member, msg)
	messages.ReplyEphemeral("Updated log channel to "+channel.Name, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating log channel select")
		err = messages.UpdateChannelMessage(
			ctx,
			b,
			selectLogChannelContents,
			messages.AdminSelectLogChannel,
			thisChannel,
		)
		if err != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select log channel message after interaction")
		}
	}()
	return nil
}
