package directmessages

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"

	"github.com/pkg/errors"
)

func getValidInvite(
	ctx context.Context,
	tx db.SafeTX,
	inviteIDstr string,
	player *models.Player,
) (*models.PlayerTeamInvite, error) {
	inviteID, err := strconv.ParseUint(inviteIDstr, 10, 0)
	if err != nil {
		return nil, errors.Wrap(err, "strconv.ParseUint")
	}
	invite, err := models.GetPlayerTeamInvite(ctx, tx, uint32(inviteID))
	if err != nil {
		return nil, errors.Wrap(err, "models.GetPlayerTeamInvite")
	}
	if invite == nil {
		return nil, errors.New("Invalid invite:This invite is no longer valid")
	}
	if invite.PlayerID != player.ID {
		return nil, errors.New("Invalid invite:This invite is not for you")
	}
	if invite.Status != nil {
		return nil, errors.New("Invalid invite:This invite is not pending")
	}
	return invite, nil
}
