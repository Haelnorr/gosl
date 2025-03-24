package directmessages

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handlerRefreshTeamPanel(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.SilentAcknowledge(i, ack)
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	pt, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	if pt == nil {
		return b.Error("Failed to refresh", "You are not currently on a team", i, *ack)
	}
	team, err := models.GetTeamByID(ctx, tx, pt.TeamID)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}

	if team.ManagerID == player.ID {
		updateTeamManagerPanel(ctx, tx, b, team, i.Message.ID, i.User.ID)
	} else {
		updateTeamPlayerPanel(ctx, tx, b, team, i.Message.ID, i.User.ID, false)
	}

	return nil
}
