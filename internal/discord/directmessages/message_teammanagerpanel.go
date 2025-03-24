package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func TeamManagerComponents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
) (*bot.MessageContents, error) {
	canRegister := true
	canInvite := true
	cantRegisterReason := "\nTo register, please complete the following:  "
	// Get current and invited players
	now := time.Now()
	currentPlayers, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return nil, errors.Wrap(err, "team.Players")
	}
	invitedPlayers, err := team.InvitedPlayers(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "team.InvitedPlayers")
	}
	if len(*currentPlayers) < 3 {
		canRegister = false
		cantRegisterReason = cantRegisterReason + "\n - Have at least 3 players"
	}
	if len(*currentPlayers) == 5 {
		canInvite = false
	}
	if team.Color == 0x181825 {
		canRegister = false
		cantRegisterReason = cantRegisterReason + "\n - Set a team color"
	}
	if team.Logo == "" {
		canRegister = false
		cantRegisterReason = cantRegisterReason + "\n - Upload a logo"
	}

	playersmsg := teamCurrentPlayersMsg(team, currentPlayers, invitedPlayers)

	// Get team registration status
	teamReg, err := team.RegistrationStatus(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "team.RegistrationStatus")
	}
	regMsg := ""
	if teamReg == nil {
		currentSeason, err := models.GetActiveSeason(ctx, tx)
		if err != nil {
			return nil, errors.Wrap(err, "models.GetActiveSeason")
		}
		if currentSeason == nil {
			canRegister = false
			cantRegisterReason = "\nThere is no active season right now"
		}
		if !currentSeason.RegistrationOpen {
			canRegister = false
			cantRegisterReason = "\nRegistration is currently closed"
		}
		regMsg = "Not currently registered"
		regMsg = regMsg + cantRegisterReason
	} else {
		canRegister = false
		status := ""
		league := fmt.Sprintf("\n__Preferred League:__ %s", teamReg.PreferredLeague)
		if teamReg.Approved == nil {
			status = fmt.Sprintf(
				"\n__Status:__ Pending approval for %s",
				teamReg.SeasonName,
			)
		} else {
			if teamReg.Placed == 0 {
				status = fmt.Sprintf(
					"\n__Status:__ Pending placement for %s",
					teamReg.SeasonName,
				)
			} else {
				status = fmt.Sprintf(
					"\n__Status:__ Placed in %s for %s",
					teamReg.PlacedLeagueName,
					teamReg.SeasonName,
				)
			}
		}
		regMsg = status + league
	}

	embed := &discordgo.MessageEmbed{
		Color: team.Color,
		Author: &discordgo.MessageEmbedAuthor{
			Name:    fmt.Sprintf("%s (%s)", team.Name, team.Abbreviation),
			IconURL: team.Logo,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Players:",
				Value:  playersmsg,
				Inline: false,
			},
			{
				Name:   "Registration:",
				Value:  regMsg,
				Inline: false,
			},
			{
				Name: "How to use:",
				Value: `
*Invite Players - Select from the list of eligible players to invite*
*Remove Players - Remove individual players from the team*
*Disband Team - Remove **ALL** players from the team, including yourself (you will be able to rejoin later if you want)*
*Register Team - Select your preferred league and register to play in the current season!*
*To upload a logo, use the **/uploadlogo** command*
`,
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: "invite_players_button",
					Label:    "Invite Players",
					Disabled: !canInvite,
				},
				&discordgo.Button{
					CustomID: "remove_players_button",
					Label:    "Remove Players",
					Style:    discordgo.DangerButton,
				},
				&discordgo.Button{
					CustomID: "disband_team_button",
					Label:    "Disband Team",
					Style:    discordgo.DangerButton,
				},
				&discordgo.Button{
					CustomID: "set_color_button",
					Label:    "Set Team Color",
				},
				&discordgo.Button{
					CustomID: "register_team_button",
					Label:    "Register Team",
					Style:    discordgo.SuccessButton,
					Disabled: !canRegister,
				},
			},
		},
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: "refresh_team_panel",
					Label:    "Refresh",
				},
			},
		},
	}
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}
