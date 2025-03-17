package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleAcceptInvite(
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
			expireInvite(b, i.Message.ID, i.User.ID, invite)
			return b.Error("Failed to accept invite", errmsg, i, *ack)
		}
		return errors.Wrap(err, "getValidInvite")
	}
	if invite.Approved != nil && *invite.Approved == 0 {
		expireInvite(b, i.Message.ID, i.User.ID, invite)
		return b.Error("Failed to accept invite", "This invite has been denied by staff", i, *ack)
	}
	currentTeam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	if currentTeam != nil {
		return b.Error("Failed to accept invite", "You are already on a team", i, *ack)
	}
	team, err := models.GetTeamByID(ctx, tx, invite.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}
	now := time.Now()
	currentPlayers, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return errors.Wrap(err, "team.Players")
	}
	if len(*currentPlayers) > 5 {
		return b.Error("Failed to accept invite", "Team is at max player count", i, *ack)
	}
	err = invite.Accept(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "invite.Accept")
	}
	resultMsg := ""
	managerMsg := ""
	if invite.Approved != nil && *invite.Approved == 1 {
		err = player.JoinTeam(ctx, tx, team.ID)
		if err != nil {
			return errors.Wrap(err, "player.JoinTeam")
		}
		resultMsg = fmt.Sprintf("You have joined %s!", team.Name)
		managerMsg = fmt.Sprintf("%s has joined %s!", player.Name, team.Name)
	} else {
		resultMsg = fmt.Sprintf(
			"You have accepted the invite to join %s and are awaiting staff approval",
			team.Name)
		managerMsg = fmt.Sprintf(
			"%s has accepted the invite to join %s and is awaiting staff approval",
			player.Name, team.Name)
	}
	expireInvite(b, i.Message.ID, i.User.ID, invite)
	go func() {
		manager, err := team.GetManager(ctx, tx)
		if err != nil {
			b.Logger.Warn().Err(err).Msg("Failed to get team manager")
			return
		}
		err = b.SendDirectMessage("Invite accepted", managerMsg, manager.DiscordID)
		if err != nil {
			b.Logger.Warn().Err(err).Msg("Failed to notify team manager of invite acceptance")
			return
		}
		updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, manager.DiscordID)
	}()
	err = b.FollowUp(resultMsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

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
		go func() {
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
	}
	err = b.FollowUp(resultMsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

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
	team, err := checkPlayerIsManager(ctx, tx, i.User.ID)
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

func expireInvite(
	b *bot.Bot,
	inviteMsgID string,
	userID string,
	invite *models.PlayerTeamInvite,
) {
	invMsg, err := b.GetDirectMessage(inviteMsgID, userID, "Team invite", 0, false)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to get direct message")
		return
	}
	contents, err := TeamInviteComponents(b, invite, "")
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to get team invite components")
		return
	}
	err = invMsg.Expire(contents)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to expire invite message")
		return
	}
}
