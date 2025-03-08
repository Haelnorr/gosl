package managerchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"sync"

	"github.com/pkg/errors"
)

const (
	managerChannelName string = "gosl-bot-leaguemanager"
)

// Setup the manager channel
func Setup(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *util.Bot,
) {
	defer wg.Done()
	b.Logger.Debug().Msg("Setting up manager channel")
	channelID, err := channels.Setup(ctx, b, channels.PurposeManager, managerChannelName)
	if err != nil {
		errch <- errors.Wrap(err, "channels.Setup (manager channel)")
		return
	}

	b.Logger.Info().Str("channel_id", channelID).Msg("Manager channel is ready")

	err = updateMessages(ctx, channelID, b)
	if err != nil {
		errch <- errors.Wrap(err, "updateMessages (manager channel)")
		return
	}
	b.Session.AddHandler(handleInteractions(ctx, b))
	b.Logger.Info().Msg("Manager channel setup complete")
}
