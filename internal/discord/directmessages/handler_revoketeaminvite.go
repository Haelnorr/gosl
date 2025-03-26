package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/util"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRevokeInvite(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	inviteIDstr string,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	_, team, err := util.CheckPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}
	inviteID, err := strconv.ParseUint(inviteIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}

	invite, err := models.GetPlayerTeamInvite(ctx, tx, uint32(inviteID))
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerTeamInvite")
	}
	if team.ID != invite.TeamID {
		return b.Error("Failed to revoke invite", "That invite is not for this team", i, *ack)
	}
	err = team.RevokeInvite(ctx, tx, invite.PlayerID)
	if err != nil {
		return errors.Wrap(err, "team.RevokeInvite")
	}

	updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, i.User.ID)

	err = b.FollowUp(fmt.Sprintf("The invite to %s to join %s has been revoked",
		invite.PlayerName, team.Name), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
