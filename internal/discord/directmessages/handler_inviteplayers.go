package directmessages

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleInvitePlayersInteraction(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	team, err := checkPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			b.Error("Interaction failed", err.Error(), i, *ack)
			return nil
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}

	contents, err := invitePlayersComponents(ctx, tx, team, i.Message.ID)
	if err != nil {
		if err.Error() == "No eligible players" {
			err := b.FollowUp("No eligible players available to invite", i)
			if err != nil {
				return errors.Wrap(err, "b.FollowUp")
			}
			return nil
		}
		return errors.Wrap(err, "invitePlayersComponents")
	}
	err = b.FollowUpComplex(contents, i, 60*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}

func handleInviteSelectedPlayersInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	team, err := checkPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			b.Error("Interaction failed", err.Error(), i, *ack)
			return nil
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}

	playerIDs := i.MessageComponentData().Values
	for _, playerID := range playerIDs {
		player, err := models.GetPlayerByDiscordID(ctx, tx, playerID)
		if err != nil {
			return errors.Wrap(err, "models.GetPlayerByDiscordID")
		}
		if player == nil {
			return errors.New("Player doesn't exist")
		}
		invite, err := team.InvitePlayer(ctx, tx, player.ID)
		if err != nil {
			return errors.Wrap(err, "team.InvitePlayer")
		}
		invMsg := bot.NewDirectMessage("Team invite", player.DiscordID, 0, false, b)
		contents, err := TeamInviteComponents(b, invite, panelMsgID)
		if err != nil {
			return errors.Wrap(err, "TeamInviteComponents")
		}
		err = invMsg.Send(contents)
		if err != nil {
			return errors.Wrap(err, "invMsg.Send")
		}
		// TODO: if invite needs approval, send to staff panel
	}

	updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, i.User.ID)

	err = b.FollowUp("Players invited", i)
	if err != nil {
		return errors.Wrap(err, "")
	}

	return nil
}
