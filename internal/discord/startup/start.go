package startup

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/adminchannel"
	"gosl/internal/discord/channels/loggingchannel"
	"gosl/internal/discord/channels/managerchannel"
	"gosl/internal/discord/commands"
	"gosl/internal/models"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Start the bot
func Start(ctx context.Context, b *bot.Bot) error {
	starttime := time.Now()
	err := b.Session.Open()
	if err != nil {
		return errors.Wrap(err, "b.session.Open")
	}

	// Setup log channel first so startup issues can be reported in discord
	b.AddChannel(logchannel.LogChannel)
	err = b.Channels[models.ChannelLog].Setup(ctx)
	if err != nil {
		return errors.Wrap(err, "Channel.Setup (LogChannel)")
	}
	// err = logchannel.Setup(ctx, b)
	// if err != nil {
	// 	return errors.Wrap(err, "b.setupLogChannel")
	// }

	// Do other setup concurrently to reduce startup time
	var wg sync.WaitGroup
	errch := make(chan error)

	// Add the setup commands here
	setups := []bot.SetupFunc{
		commands.Setup,
		adminchannel.Setup,
		managerchannel.Setup,
	}

	// Run all the setup commands
	for _, setup := range setups {
		wg.Add(1)
		go setup(&wg, errch, ctx, b)
	}

	go func() {
		wg.Wait()
		close(errch)
	}()

	var boterrors []error
	for err := range errch {
		if err != nil {
			b.Logger.Error().Err(err).Msg("Error in bot startup")
			boterrors = append(boterrors, err)
		}
	}
	if len(boterrors) > 0 {
		msg := "\n"
		for _, err := range boterrors {
			msg = msg + "Error: " + err.Error() + "\n\n"
		}
		err = errors.New(msg)
		b.Log().Error("**Error(s) during bot startup**", err)
		return errors.New("Error(s) during bot startup")
	}
	b.Logger.Info().Dur("startup_time", time.Since(starttime)).Msg("Bot startup complete!")
	b.Log().Info("Bot startup complete")
	return nil
}

// Stop the bot
func Stop(b *bot.Bot) error {
	err := b.Session.Close()
	if err != nil {
		return err
	}
	return nil
}
