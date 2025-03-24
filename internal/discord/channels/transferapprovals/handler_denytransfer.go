package transferapprovals

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRejectTransfer(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	ptiIDstr string,
) error {
	b.Acknowledge(i, ack)
	ptiID, err := strconv.ParseUint(ptiIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	pti, err := models.GetPlayerTeamInvite(ctx, tx, uint32(ptiID))
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerTeamInvite")
	}
	if pti.Approved != nil {
		updateRequestMsg(ctx, tx, b, i, pti, true)
		return b.Error("Failed to approve transfer", "Transfer is not pending", i, *ack)
	}
	if pti.Status != nil && *pti.Status == 0 {
		updateRequestMsg(ctx, tx, b, i, pti, true)
		return b.FollowUp("Transfer was rejected by the player", i)
	}
	team, err := models.GetTeamByID(ctx, tx, pti.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}

	err = pti.Deny(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "pti.Deny")
	}
	playermsg := fmt.Sprintf("Your invite to join %s has been denied.", pti.TeamName)
	managermsg := fmt.Sprintf(
		"The invite for %s to join %s has been denied.",
		pti.PlayerName, pti.TeamName)

	if pti.Status == nil || *pti.Status == 1 {
		player, err := models.GetPlayerByID(ctx, tx, pti.PlayerID)
		if err != nil {
			return errors.Wrap(err, "models.GetPlayerByID")
		}
		err = b.SendDirectMessage("Team Invite Denied", playermsg, player.DiscordID)
		if err != nil {
			return errors.Wrap(err, "b.SendDirectMessage")
		}
		manager, err := models.GetPlayerByID(ctx, tx, team.ManagerID)
		if err != nil {
			return errors.Wrap(err, "models.GetPlayerByID")
		}
		err = b.SendDirectMessage("Team Invite Denied", managermsg, manager.DiscordID)
		if err != nil {
			return errors.Wrap(err, "b.SendDirectMessage")
		}
	}

	b.Log().UserEvent(i.Member, managermsg)
	updateRequestMsg(ctx, tx, b, i, pti, true)

	// TODO: add transfer window handling
	err = b.FollowUp(managermsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
