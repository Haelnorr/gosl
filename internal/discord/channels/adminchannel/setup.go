package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"sync"

	"github.com/pkg/errors"
)

// Setup the admin channel
func Setup(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *bot.Bot,
) {
	defer wg.Done()
	channel := &bot.Channel{
		Purpose: models.ChannelAdmin,
		Name:    "gosl-bot-admin",
		Label:   "Admin channel",
		Handler: handleInteractions(ctx, b),
	}
	err := b.AddChannel(channel)
	if err != nil {
		errch <- errors.Wrap(err, "b.AddChannel")
		return
	}
	err = channel.Setup(ctx)
	if err != nil {
		errch <- errors.Wrap(err, "channel.Setup")
		return
	}

	// register all the messages
	var errs []error
	errs = append(errs, channel.RegisterMessage(selectLogChannel))
	errs = append(errs, channel.RegisterMessage(selectAdminRoles))
	errs = append(errs, channel.RegisterMessage(selectManagerRoles))

	// check for any errors setting up messages and return if any occured
	hadErr := false
	for _, err := range errs {
		if err != nil {
			errch <- errors.Wrap(err, "channel.RegisterMessage")
			hadErr = true
		}
	}
	if hadErr {
		return
	}
	channel.SetupMessages(ctx, errch)
}
