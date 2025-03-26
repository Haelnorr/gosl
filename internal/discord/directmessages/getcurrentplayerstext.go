package directmessages

import (
	"fmt"
	"gosl/internal/models"
	"sort"
)

// Generates a string with the list of players
func teamCurrentPlayersMsg(
	team *models.Team,
	currentPlayers *[]models.Player,
	invitedPlayers *[]models.PlayerTeamInvite,
) string {
	playersmsg := ""
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
	return playersmsg
}
