package util

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Takes a users Discord ID, checks if the player is registered, and if they
// have a current team, and if they are the manager of that team.
// Validation errors will result in a new error with prefix "VE:"
// Returns the player and the team
func CheckPlayerIsManager(
	ctx context.Context,
	tx db.SafeTX,
	discordID string,
) (*models.Player, *models.Team, error) {
	player, err := models.GetPlayerByDiscordID(ctx, tx, discordID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player == nil {
		return nil, nil, errors.New("VE: Not registered as a player")
	}
	playerteam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "player.CurrentTeam")
	}
	if playerteam == nil {
		return nil, nil, errors.New("VE: Not currently on a team")
	}
	team, err := models.GetTeamByID(ctx, tx, playerteam.TeamID)
	if err != nil {
		return nil, nil, errors.Wrap(err, "models.GetTeamByID")
	}
	if team.ManagerID != player.ID {
		return nil, nil, errors.New("VE: Not currently manager of this team")
	}

	return player, team, nil
}
