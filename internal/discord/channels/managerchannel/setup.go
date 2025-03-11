package managerchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"sync"
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
	channels.Setup(
		errch, ctx, b,
		channels.PurposeManager,
		managerChannelName,
		updateMessages,
		handleInteractions(ctx, b),
	)
}
