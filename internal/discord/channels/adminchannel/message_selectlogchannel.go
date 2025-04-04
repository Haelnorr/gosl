package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"

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
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select log channel components")
	logChannelID, err := models.GetChannel(ctx, tx, models.ChannelLog)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelForPurpose")
	}
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
		Components: components.ActionRow(components.ChannelSelect(
			"log_channel_select",
			"Select the channel for log output",
			defaultValues,
			1,
			1,
			[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
		)),
	}
	return contents, nil
}
