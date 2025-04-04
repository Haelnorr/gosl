package teamrosters

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRefresh(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.SilentAcknowledge(i, ack)
	err := UpdateTeamRosters(ctx, tx, b)
	if err != nil {
		return errors.Wrap(err, "UpdateTeamRosters")
	}
	return nil
}
