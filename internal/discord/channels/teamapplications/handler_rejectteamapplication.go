package teamapplications

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

func handleRejectTeamApplication(
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

	err = app.Reject(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "app.Reject")
	}
	err = b.SendDirectMessage("Team Application Rejected",
		fmt.Sprintf("Your application for %s to play in %s has been rejected",
			app.TeamName, app.SeasonName),
		app.ManagerID,
	)
	if err != nil {
		return errors.Wrap(err, "b.SendDirectMessage")
	}

	err = updateAppMsg(ctx, tx, b, i, app, true)
	if err != nil {
		return errors.Wrap(err, "updateAppMsg")
	}
	msg := fmt.Sprintf("Application from %s rejected", app.TeamName)
	b.Log().UserEvent(i.Member, msg)
	return b.FollowUp(msg, i)
}
