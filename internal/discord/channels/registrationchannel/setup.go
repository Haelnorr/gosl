package registrationchannel

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
		Purpose: models.ChannelRegistration,
		Label:   "Registration channel",
		Handler: handleInteractions(ctx, b),
	}
	err := b.AddChannel(channel)
	if err != nil {
		errch <- errors.Wrap(err, "b.AddChannel")
		return
	}
	err = channel.Setup(ctx, false)
	if err != nil {
		errch <- errors.Wrap(err, "channel.Setup")
		return
	}

	// register all the messages
	var errs []error
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
	var mwg sync.WaitGroup
	mwg.Add(1)
	channel.SetupMessages(ctx, &mwg, errch)
}
