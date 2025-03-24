package freeagentapplications

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/teamrosters"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleApproveFreeAgentApplication(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	appIDstr string,
) error {
	b.Acknowledge(i, ack)
	appID, err := strconv.ParseUint(appIDstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	app, err := models.GetFreeAgentRegistration(ctx, tx, uint32(appID))
	if err != nil {
		return errors.Wrap(err, "models.GetFreeAgentRegistration")
	}
	err = app.Approve(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "app.Approve")
	}
	player, err := models.GetPlayerByID(ctx, tx, app.PlayerID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByID")
	}
	err = b.SendDirectMessage("Free Agent Application Approved",
		fmt.Sprintf("Your application to play in %s as a Free Agent has been approved",
			app.SeasonName),
		player.DiscordID,
	)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}
	err = updateAppMsg(ctx, tx, b, i, app, false)
	if err != nil {
		return errors.Wrap(err, "updateAppMsg")
	}
	err = teamrosters.UpdateTeamRosters(ctx, b)
	if err != nil {
		return errors.Wrap(err, "teamrosters.UpdateTeamRosters")
	}

	msg := fmt.Sprintf("Application from %s approved", app.PlayerName)
	b.Log().UserEvent(i.Member, msg)
	return b.FollowUp(msg, i)
}
