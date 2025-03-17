package directmessages

import (
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

func removePlayersComponents(
	team *models.Team,
	currentPlayers *[]models.Player,
	invitedPlayers *[]models.PlayerTeamInvite,
	messageID string,
) *bot.MessageContents {
	currentPlayerButtons := []discordgo.MessageComponent{}
	invitedPlayerButtons := []discordgo.MessageComponent{}
	for _, player := range *currentPlayers {
		if player.ID != team.ManagerID {
			currentPlayerButtons = append(currentPlayerButtons, &discordgo.Button{
				CustomID: fmt.Sprintf("remove_player_%v_%s", player.ID, messageID),
				Label:    fmt.Sprintf("Remove %s", player.Name),
				Style:    discordgo.DangerButton,
			})
		}
	}
	for _, invite := range *invitedPlayers {
		invitedPlayerButtons = append(invitedPlayerButtons, &discordgo.Button{
			CustomID: fmt.Sprintf("revoke_invite_%v_%s", invite.ID, messageID),
			Label:    fmt.Sprintf("Revoke invite for %s", invite.PlayerName),
			Style:    discordgo.DangerButton,
		})
	}
	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: fmt.Sprintf("Remove players from %s (%s)", team.Name, team.Abbreviation),
				Value: `
Click on the buttons below to remove players from the team, or revoke their invites.
`,
				Inline: false,
			},
		},
	}
	msgcomps := []discordgo.MessageComponent{}
	if len(currentPlayerButtons) > 0 {
		msgcomps = append(msgcomps, &discordgo.ActionsRow{
			Components: currentPlayerButtons,
		})
	}
	if len(invitedPlayerButtons) > 0 {
		msgcomps = append(msgcomps, &discordgo.ActionsRow{
			Components: invitedPlayerButtons,
		})
	}
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents
}
