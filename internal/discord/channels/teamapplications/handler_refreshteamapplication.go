package teamapplications

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRefreshTeamApplication(
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
	err = updateAppMsg(ctx, tx, b, i, app, false)
	if err != nil {
		return errors.Wrap(err, "updateAppMsg")
	}
	return b.FollowUp("Refreshed", i)
}
