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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "selectChannelsContents()")
	if err != nil {
		return nil, errors.Wrap(err, "conn.RBegin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select channel components")
	regChannelSelect, err := getChannelSelect(ctx, tx, models.ChannelRegistration,
		"registration_channel_select", "Player/Team/Free Agent Registrations", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelSelect")
	}
	teamAppChannelSelect, err := getChannelSelect(ctx, tx, models.ChannelTeamApplications,
		"team_application_channel_select", "Team Applications", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelSelect")
	}
	freeAgentAppChannelSelect, err := getChannelSelect(ctx, tx, models.ChannelFreeAgentApplications,
		"freeagent_application_channel_select", "Free Agent Applications", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelSelect")
	}
	transferApprovalsChannelSelect, err := getChannelSelect(ctx, tx, models.ChannelTransferApprovals,
		"transfer_approval_channel_select", "Transfer Approvals", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelSelect")
	}
	teamRostersChannelSelect, err := getChannelSelect(ctx, tx, models.ChannelTeamRosters,
		"team_rosters_channel_select", "Team/Free Agent Rosters", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getChannelSelect")
	}
	tx.Commit()

	embed := &discordgo.MessageEmbed{
		Title: "Select Channels",
		Description: `
**Registrations:**
Channel for players to create teams and register to play in OSL

**Team Applications:**
Channel for viewing and actioning team applications"

**Free Agent Applications:**
Channel for viewing and actioning free agent applications"

**Transfer Approvals:**
Channel for viewing and actioning transfer applications"

**Team Rosters:**
Channel for viewing Team Rosters and Free Agents"
`,
		Color: 0x00ff00, // Green color
	}

	comps := regChannelSelect
	comps = append(comps, teamAppChannelSelect...)
	comps = append(comps, freeAgentAppChannelSelect...)
	comps = append(comps, transferApprovalsChannelSelect...)
	comps = append(comps, teamRostersChannelSelect...)
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: comps,
	}
	return contents, nil
}

func getChannelSelect(
	ctx context.Context,
	tx db.SafeTX,
	channelPurpose uint16,
	customID string,
	placeholder string,
	minValues int,
	maxValues int,
) ([]discordgo.MessageComponent, error) {
	channelID, err := models.GetChannel(ctx, tx, channelPurpose)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	var defaults []discordgo.SelectMenuDefaultValue
	if channelID != "" {
		defaults = append(defaults, discordgo.SelectMenuDefaultValue{
			ID:   channelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	return components.ActionRow(components.ChannelSelect(
		customID,
		placeholder,
		defaults,
		minValues,
		maxValues,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)), nil
}
