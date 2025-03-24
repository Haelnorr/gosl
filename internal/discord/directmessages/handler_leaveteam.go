package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleLeaveTeamButton(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	// TODO: fail if team currently registered for a season? ask LCs
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	pt, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	if pt == nil {
		return b.Error("Failed to leave team", "You are not on a team", i, *ack)
	}
	team, err := models.GetTeamByID(ctx, tx, pt.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}
	contents := confirmLeaveTeamComponents(team, i.Message.ID)
	err = b.FollowUpComplex(contents, i, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}

func handleLeaveTeamConfirm(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	pt, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	if pt == nil {
		return b.Error("Failed to leave team", "You are not on a team", i, *ack)
	}
	team, err := models.GetTeamByID(ctx, tx, pt.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}
	err = player.LeaveTeam(ctx, tx, team.ID)
	if err != nil {
		return errors.Wrap(err, "player.LeaveTeam")
	}
	updateTeamPlayerPanel(ctx, tx, b, team, panelMsgID, i.User.ID, true)
	err = b.FollowUp(fmt.Sprintf("You have left %s", team.Name), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
