package adminchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"sync"

	"github.com/pkg/errors"
)

const (
	adminChannelName string = "gosl-bot-admin"
)

// Setup the admin channel
func Setup(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *util.Bot,
) {
	defer wg.Done()
	b.Logger.Debug().Msg("Setting up admin channel")

	channelID, err := channels.Setup(ctx, b, channels.PurposeAdmin, adminChannelName)
	if err != nil {
		errch <- errors.Wrap(err, "channels.Setup (admin channel)")
		return
	}

	b.Logger.Info().Str("channel_id", channelID).Msg("Admin channel is ready")

	err = updateMessages(ctx, channelID, b)
	if err != nil {
		errch <- errors.Wrap(err, "updateMessages (admin channel)")
		return
	}
	b.Session.AddHandler(handleInteractions(ctx, b))
	b.Logger.Info().Msg("Admin channel setup complete")
}
