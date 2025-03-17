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

var selectChannels = &bot.Message{
	Label:       "Select Channels",
	Purpose:     models.MsgSelectChannels,
	GetContents: selectChannelsContents,
}

// Get the message contents for the select channels message
func selectChannelsContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select channels message")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "conn.RBegin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select channel components")
	regChannelID, err := models.GetChannel(ctx, tx, models.ChannelRegistration)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	appChannelID, err := models.GetChannel(ctx, tx, models.ChannelRegistrationApproval)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	tx.Commit()

	var registrationDefaults []discordgo.SelectMenuDefaultValue
	if regChannelID != "" {
		registrationDefaults = append(registrationDefaults, discordgo.SelectMenuDefaultValue{
			ID:   regChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	var applicationsDefaults []discordgo.SelectMenuDefaultValue
	if appChannelID != "" {
		applicationsDefaults = append(applicationsDefaults, discordgo.SelectMenuDefaultValue{
			ID:   appChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	embed := &discordgo.MessageEmbed{
		Title: "Select Channels",
		Description: `
**Player/Team/Free Agent Registrations:**
Channel for players create teams and register to play in OSL

**Team/Free Agent Applications:**
Channel for viewing and actioning registration applications"
`,
		Color: 0x00ff00, // Green color
	}
	comps := components.ChannelSelect(
		"registration_channel_select",
		"Player/Team/Free Agent Registrations",
		registrationDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)
	comps = append(comps, components.ChannelSelect(
		"application_channel_select",
		"Team/Free Agent Applications",
		applicationsDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)...)
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: comps,
	}
	return contents, nil
}
