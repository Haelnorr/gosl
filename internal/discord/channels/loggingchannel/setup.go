package logchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"

	"github.com/pkg/errors"
)

const (
	logChannelName string = "gosl-bot-log"
)

// Setup the logging channel
func Setup(
	ctx context.Context,
	b *util.Bot,
) error {
	b.Logger.Debug().Msg("Setting up log channel")
	channelID, err := channels.Setup(ctx, b, channels.PurposeLog, logChannelName)
	if err != nil {
		return errors.Wrap(err, "channels.Setup (log channel)")
	}
	b.Logchannel = channelID
	b.Logger.Info().Msg("Log channel setup complete")
	return nil
}
