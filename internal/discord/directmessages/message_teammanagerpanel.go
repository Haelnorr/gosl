package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"sort"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// TODO: interactions: register team
// TODO: figure out logo uploading
func TeamManagerComponents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
) (*bot.MessageContents, error) {
	canRegister := true
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
	playersmsg := "**Players:**"
	sort.SliceStable(*currentPlayers, func(i, j int) bool {
		if (*currentPlayers)[i].ID == team.ManagerID {
			return true
		}
		if (*currentPlayers)[j].ID == team.ManagerID {
			return false
		}
		return false
	})
	for _, player := range *currentPlayers {
		if player.ID == team.ManagerID {
			playersmsg = playersmsg + "\n%s (Manager)"
		} else {
			playersmsg = playersmsg + "\n%s"
		}
		playersmsg = fmt.Sprintf(playersmsg, player.Name)
	}
	for _, player := range *invitedPlayers {
		playersmsg = playersmsg + "\n%s (%s)"
		if player.Status == nil {
			playersmsg = fmt.Sprintf(playersmsg, player.PlayerName, "Invited")
		} else if player.Approved == nil {
			playersmsg = fmt.Sprintf(playersmsg, player.PlayerName, "Pending approval")
		}
	}
	if len(*currentPlayers) < 3 {
		canRegister = false
		cantRegisterReason = cantRegisterReason + "\n - Have at least 3 players"
	}
	if team.Color == 0x181825 {
		canRegister = false
		cantRegisterReason = cantRegisterReason + "\n - Set a team color"
	}
	// TODO: team logo check

	// Get team registration status
	teamReg, err := team.RegistrationStatus(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "team.RegistrationStatus")
	}
	regMsg := "**Registration:** %s"
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
		regMsg = fmt.Sprintf(regMsg, "Not currently registered")
		regMsg = regMsg + cantRegisterReason
	} else {
		canRegister = false
		status := ""
		league := fmt.Sprintf("\n__Preferred League:__ %s", teamReg.PreferredLeague)
		if teamReg.Approved == nil {
			status = fmt.Sprintf("\n__Status:__ Pending approval for %s", teamReg.SeasonName)
		} else {
			if teamReg.Placed == 0 {
				status = fmt.Sprintf("\n__Status:__ Pending placement for %s", teamReg.SeasonName)
			}
		}
		regMsg = fmt.Sprintf(regMsg, status+league)
	}

	embed := &discordgo.MessageEmbed{
		Color: team.Color,
		// USE FOR TEAM WITH LOGO SUBMITTED ALREADY
		// Author: &discordgo.MessageEmbedAuthor{
		// 	Name:    teamName,
		// 	IconURL:.Avatar,
		// },
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: fmt.Sprintf("%s (%s)", team.Name, team.Abbreviation),
				Value: fmt.Sprintf(`
%s

%s

*Invite Players - Select from the list of eligible players to invite*
*Remove Players - Remove individual players from the team*
*Disband Team - Remove **ALL** players from the team, including yourself (you will be able to rejoin later if you want)*
*Register Team - Select your preferred league and register to play in the current season!*
`, playersmsg, regMsg),
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
	}
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}

func updateTeamManagerPanel(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
	panelMsgID string,
	userID string,
) {
	panelMsg, err := b.GetDirectMessage(
		panelMsgID,
		userID,
		"Team Manager Panel",
		5*time.Minute,
		false,
	)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "b.GetDirectMessage")).
			Msg("Failed to update team manager panel")
		return
	}
	contents, err := TeamManagerComponents(ctx, tx, b, team)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "TeamManagerComponents")).
			Msg("Failed to update team manager panel")
		return
	}
	err = panelMsg.Update(contents)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "panelMsg.Update")).
			Msg("Failed to update team manager panel")
		return
	}
}
