package transferapprovals

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/teamrosters"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleApproveTransfer(
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
	now := time.Now()
	players, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return errors.Wrap(err, "team.Players")
	}
	if len(*players) == 5 {
		return b.Error("Failed to approve transfer", "Team has 5 players already", i, *ack)
	}

	player, err := models.GetPlayerByID(ctx, tx, pti.PlayerID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByID")
	}

	err = pti.Approve(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "pti.Approve")
	}
	playermsg := fmt.Sprintf(
		"Your invite to join %s has been approved. You have not yet accepted",
		pti.TeamName)
	managermsg := fmt.Sprintf(
		"The invite for %s to join %s has been approved. The player has not yet accepted",
		pti.PlayerName, pti.TeamName)
	if pti.Status != nil && *pti.Status == 1 {
		currentTeam, err := player.CurrentTeam(ctx, tx)
		if err != nil {
			return errors.Wrap(err, "player.CurrentTeam")
		}
		if currentTeam != nil {
			// We manually rollback here to force the approval to revert
			// without returning a system error
			tx.Rollback()
			return b.Error("Failed to approve transfer", "Player is already in a team", i, *ack)
		}
		err = player.JoinTeam(ctx, tx, pti.TeamID)
		if err != nil {
			return errors.Wrap(err, "player.JoinTeam")
		}
		playermsg = fmt.Sprintf(
			"Your invite to join %s has been approved. You have now joined the team",
			pti.TeamName)
		managermsg = fmt.Sprintf(
			"The invite for %s to join %s has been approved. The player has joined the team",
			pti.PlayerName, pti.TeamName)
	}
	err = b.SendDirectMessage("Team Invite Approved", playermsg, player.DiscordID)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}
	manager, err := models.GetPlayerByID(ctx, tx, team.ManagerID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByID")
	}
	err = b.SendDirectMessage("Team Invite Approved", managermsg, manager.DiscordID)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}

	b.Log().UserEvent(i.Member, managermsg)
	updateRequestMsg(ctx, tx, b, i, pti, true)
	err = teamrosters.UpdateTeamRosters(ctx, b)
	if err != nil {
		return errors.Wrap(err, "teamrosters.UpdateTeamRosters")
	}

	// TODO: add transfer window handling
	err = b.FollowUp(managermsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
