package adminchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"sync"
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
	channels.Setup(
		errch, ctx, b,
		channels.PurposeAdmin,
		adminChannelName,
		updateMessages,
		handleInteractions(ctx, b),
	)
}
