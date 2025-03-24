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

func TeamPlayerComponents(
	ctx context.Context,
	tx db.SafeTX,
	team *models.Team,
) (*bot.MessageContents, error) {
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
	playersmsg := teamCurrentPlayersMsg(team, currentPlayers, invitedPlayers)
	// Get team registration status
	teamReg, err := team.RegistrationStatus(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "team.RegistrationStatus")
	}
	regMsg := ""
	if teamReg == nil {
		regMsg = "Not currently registered"
	} else {
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
		},
	}
	msgcomps := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.Button{
					CustomID: "refresh_team_panel",
					Label:    "Refresh",
				},
				&discordgo.Button{
					CustomID: "leave_team_button",
					Label:    "Leave Team",
					Style:    discordgo.DangerButton,
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
