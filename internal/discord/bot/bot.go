package bot

import (
	"context"
	"gosl/internal/discord/channels/adminchannel"
	"gosl/internal/discord/channels/loggingchannel"
	"gosl/internal/discord/channels/managerchannel"
	"gosl/internal/discord/commands"
	"gosl/internal/discord/util"
	"gosl/pkg/config"
	"gosl/pkg/db"
	"io/fs"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Create a new discord bot and open a session
func NewBot(
	config *config.Config,
	conn *db.SafeConn,
	logger *zerolog.Logger,
	fs *fs.FS,
) (*util.Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordBotToken)
	if err != nil {
		return nil, errors.Wrap(err, "discordgo.New")
	}
	b := &util.Bot{Session: session, Logger: logger, Files: fs, Conn: conn, Config: config}
	return b, nil
}

// Start the bot
func Start(ctx context.Context, b *util.Bot) error {
	err := b.Session.Open()
	if err != nil {
		return errors.Wrap(err, "b.session.Open")
	}

	// Setup log channel first so startup issues can be reported in discord
	err = logchannel.Setup(ctx, b)
	if err != nil {
		return errors.Wrap(err, "b.setupLogChannel")
	}

	// Do other setup concurrently to reduce startup time
	var wg sync.WaitGroup
	errch := make(chan error)

	// Add the setup commands here
	setups := []util.SetupFunc{
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
	b.Logger.Info().Msg("Bot startup complete!")
	b.Log().Info("Bot startup complete")
	return nil
}

// Stop the bot
func Stop(b *util.Bot) error {
	err := b.Session.Close()
	if err != nil {
		return err
	}
	return nil
}
