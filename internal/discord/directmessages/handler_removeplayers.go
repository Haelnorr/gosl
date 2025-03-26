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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRemovePlayersButton(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	_, team, err := util.CheckPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}
	now := time.Now()
	currentPlayers, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return errors.Wrap(err, "team.Players")
	}
	invitedPlayers, err := team.InvitedPlayers(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "team.InvitedPlayers")
	}

	contents := removePlayersComponents(team, currentPlayers, invitedPlayers, i.Message.ID)
	err = b.FollowUpComplex(contents, i, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}

	return nil
}

func handleRemovePlayerInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	playerIDstr string,
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
	playerID, err := strconv.ParseUint(playerIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	player, err := models.GetPlayerByID(ctx, tx, uint16(playerID))
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByID")
	}
	err = player.LeaveTeam(ctx, tx, team.ID)
	if err != nil {
		return errors.Wrap(err, "player.LeaveTeam")
	}
	err = b.SendDirectMessage(
		"Removed from Team",
		fmt.Sprintf("You have been removed from %s", team.Name),
		player.DiscordID,
	)
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to notify player of removal from team")
		err = nil
	}

	updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, i.User.ID)

	err = b.FollowUp(fmt.Sprintf("%s was removed from %s", player.Name, team.Name), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}

	return nil
}
