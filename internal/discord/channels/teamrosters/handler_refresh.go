package teamrosters

import (
	"context"
	"gosl/internal/discord/bot"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRefresh(
	ctx context.Context,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.SilentAcknowledge(i, ack)
	err := UpdateTeamRosters(ctx, b)
	if err != nil {
		return errors.Wrap(err, "UpdateTeamRosters")
	}
	return nil
}
