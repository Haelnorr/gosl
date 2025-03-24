package directmessages

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/pkg/errors"
)

func updateTeamManagerPanel(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
	panelMsgID string,
	userID string,
) {
	panelMsg, err := b.GetDirectMessage(
		panelMsgID,
		userID,
		"Team Manager Panel",
		5*time.Minute,
		false,
	)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "b.GetDirectMessage")).
			Msg("Failed to update team manager panel")
		return
	}
	contents, err := TeamManagerComponents(ctx, tx, b, team)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "TeamManagerComponents")).
			Msg("Failed to update team manager panel")
		return
	}
	err = panelMsg.Update(contents)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "panelMsg.Update")).
			Msg("Failed to update team manager panel")
		return
	}
}
