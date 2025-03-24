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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "selectChannelsContents()")
	if err != nil {
		return nil, errors.Wrap(err, "conn.RBegin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select channel components")
	regChannelID, err := models.GetChannel(ctx, tx, models.ChannelRegistration)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	teamAppChannelID, err := models.GetChannel(ctx, tx, models.ChannelTeamApplications)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	freeAgentAppChannelID, err := models.GetChannel(ctx, tx, models.ChannelFreeAgentApplications)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	teamRostersChannelID, err := models.GetChannel(ctx, tx, models.ChannelTeamRosters)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	transferApprovalsChannelID, err := models.GetChannel(ctx, tx, models.ChannelTransferApprovals)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetChannel")
	}
	tx.Commit()

	var registrationDefaults []discordgo.SelectMenuDefaultValue
	var teamApplicationsDefaults []discordgo.SelectMenuDefaultValue
	var freeAgentApplicationsDefaults []discordgo.SelectMenuDefaultValue
	var teamRostersDefaults []discordgo.SelectMenuDefaultValue
	var transferApprovalsDefaults []discordgo.SelectMenuDefaultValue
	if regChannelID != "" {
		registrationDefaults = append(registrationDefaults, discordgo.SelectMenuDefaultValue{
			ID:   regChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	if teamAppChannelID != "" {
		teamApplicationsDefaults = append(teamApplicationsDefaults, discordgo.SelectMenuDefaultValue{
			ID:   teamAppChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	if freeAgentAppChannelID != "" {
		freeAgentApplicationsDefaults = append(freeAgentApplicationsDefaults, discordgo.SelectMenuDefaultValue{
			ID:   freeAgentAppChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	if teamRostersChannelID != "" {
		teamRostersDefaults = append(teamRostersDefaults, discordgo.SelectMenuDefaultValue{
			ID:   teamRostersChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	if transferApprovalsChannelID != "" {
		transferApprovalsDefaults = append(transferApprovalsDefaults, discordgo.SelectMenuDefaultValue{
			ID:   transferApprovalsChannelID,
			Type: discordgo.SelectMenuDefaultValueChannel,
		})
	}
	embed := &discordgo.MessageEmbed{
		Title: "Select Channels",
		Description: `
**Player/Team/Free Agent Registrations:**
Channel for players create teams and register to play in OSL

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
	comps := components.ChannelSelect(
		"registration_channel_select",
		"Player/Team/Free Agent Registrations",
		registrationDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)
	comps = append(comps, components.ChannelSelect(
		"team_application_channel_select",
		"Team Applications",
		teamApplicationsDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)...)
	comps = append(comps, components.ChannelSelect(
		"freeagent_application_channel_select",
		"Free Agent Applications",
		freeAgentApplicationsDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)...)
	comps = append(comps, components.ChannelSelect(
		"transfer_approval_channel_select",
		"Transfer Approvals",
		transferApprovalsDefaults,
		1,
		1,
		[]discordgo.ChannelType{discordgo.ChannelTypeGuildText},
	)...)
	comps = append(comps, components.ChannelSelect(
		"team_rosters_channel_select",
		"Team/Free Agent Rosters",
		teamRostersDefaults,
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
