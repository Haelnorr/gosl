package directmessages

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/pkg/errors"
)

func updateTeamPlayerPanel(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
	panelMsgID string,
	userID string,
	locked bool,
) {
	panelMsg, err := b.GetDirectMessage(
		panelMsgID,
		userID,
		"Team Player Panel",
		5*time.Minute,
		false,
	)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "b.GetDirectMessage")).
			Msg("Failed to update team player panel")
		return
	}
	contents, err := TeamPlayerComponents(ctx, tx, team)
	if err != nil {
		b.Logger.Warn().Err(errors.Wrap(err, "TeamPlayerComponents")).
			Msg("Failed to update team player panel")
		return
	}
	if locked {
		err = panelMsg.Expire(contents)
		if err != nil {
			b.Logger.Warn().Err(errors.Wrap(err, "panelMsg.Expire")).
				Msg("Failed to update team player panel")
			return
		}
	} else {
		err = panelMsg.Update(contents)
		if err != nil {
			b.Logger.Warn().Err(errors.Wrap(err, "panelMsg.Update")).
				Msg("Failed to update team player panel")
			return
		}
	}
}
