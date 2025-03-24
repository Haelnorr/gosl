package teamapplications

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

func handleApproveTeamApplication(
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
	app, err := models.GetTeamRegistration(ctx, tx, uint16(appID))
	if err != nil {
		return errors.Wrap(err, "models.GetTeamRegistration")
	}

	err = app.Approve(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "app.Approve")
	}
	err = b.SendDirectMessage("Team Application Approved",
		fmt.Sprintf("Your application for %s to play in %s has been approved",
			app.TeamName, app.SeasonName),
		app.ManagerID,
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

	msg := fmt.Sprintf("Application from %s approved", app.TeamName)
	b.Log().UserEvent(i.Member, msg)
	return b.FollowUp(msg, i)
}
