package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRejectInvite(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	inviteIDstr string,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	invite, err := getValidInvite(ctx, tx, inviteIDstr, player)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid invite:") {
			errmsg := strings.TrimPrefix(err.Error(), "Invalid invite:")
			return b.Error("Failed to reject invite", errmsg, i, *ack)
		}
		return errors.Wrap(err, "getValidInvite")
	}
	team, err := models.GetTeamByID(ctx, tx, invite.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}
	err = invite.Reject(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "invite.Reject")
	}
	resultMsg := fmt.Sprintf("You have rejected an invite to %s!", team.Name)

	expireInvite(b, i.Message.ID, i.User.ID, invite)

	if invite.Approved == nil || *invite.Approved == 1 {
		managerMsg := fmt.Sprintf("%s has rejected your invite to %s!", player.Name, team.Name)
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			manager, err := team.GetManager(ctx, tx)
			if err != nil {
				b.Logger.Warn().Err(err).Msg("Failed to get team manager")
				return
			}
			err = b.SendDirectMessage("Invite rejected", managerMsg, manager.DiscordID)
			if err != nil {
				b.Logger.Warn().Err(err).Msg("Failed to notify team manager of invite rejected")
				return
			}
			updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, manager.DiscordID)
		}()
		wg.Wait()
	}
	err = b.FollowUp(resultMsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
