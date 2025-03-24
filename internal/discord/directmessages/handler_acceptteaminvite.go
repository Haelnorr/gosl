package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"
	"sync"
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
	if len(*currentPlayers) == 5 {
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
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
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
	wg.Wait()
	err = b.FollowUp(resultMsg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
