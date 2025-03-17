package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectLogChannel = &bot.Message{
	Label:       "Select Log Channel",
	Purpose:     models.MsgSelectLogChannel,
	GetContents: selectLogChannelContents,
}

// Get the message contents for the select log channel component
func selectLogChannelContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select log channel components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select log channel components")
	logChannelID, err := models.GetChannel(ctx, tx, models.ChannelLog)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelForPurpose")
	}
	tx.Commit()
	var defaultValues []discordgo.SelectMenuDefaultValue
	defaultValues = append(defaultValues, discordgo.SelectMenuDefaultValue{
		ID:   logChannelID,
		Type: discordgo.SelectMenuDefaultValueChannel,
	})
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title:       "Log output channel",
			Description: `Select the channel to output bot logs to`,
			Color:       0x00ff00, // Green color
		},
		Components: components.ChannelSelect(
			"log_channel_select",
			"Select the channel for log output",
			defaultValues,
			1,
			1,
			[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
		),
	}
	return contents, nil
}
