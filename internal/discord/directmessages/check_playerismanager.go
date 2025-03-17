package directmessages

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

func checkPlayerIsManager(
	ctx context.Context,
	tx db.SafeTX,
	discordID string,
) (*models.Team, error) {
	player, err := models.GetPlayerByDiscordID(ctx, tx, discordID)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player == nil {
		return nil, errors.New("VE: Not registered as a player")
	}
	playerteam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "player.CurrentTeam")
	}
	if playerteam == nil {
		return nil, errors.New("VE: Not currently on a team")
	}
	team, err := models.GetTeamByID(ctx, tx, playerteam.TeamID)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetTeamByID")
	}
	if team.ManagerID != player.ID {
		return nil, errors.New("VE: Not currently manager of this team")
	}

	return team, nil
}
