package bot

import (
	"context"
	"gosl/internal/discord/channels/adminchannel"
	logchannel "gosl/internal/discord/channels/loggingchannel"
	"gosl/internal/discord/commands"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"io/fs"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Create a new discord bot and open a session
func NewBot(
	token string,
	guildID string,
	conn *db.SafeConn,
	logger *zerolog.Logger,
	fs *fs.FS,
) (*util.Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, errors.Wrap(err, "discordgo.New")
	}
	b := &util.Bot{Session: session, Logger: logger, Files: fs, Conn: conn, GuildID: guildID}
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
	errch := make(chan error, 10)

	// Add the setup commands here
	setups := []util.SetupFunc{
		commands.Setup,
		adminchannel.Setup,
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

	hadErrors := false
	for err := range errch {
		if err != nil {
			b.Logger.Error().Err(err).Msg("Error in bot startup")
			hadErrors = true
		}
	}
	if hadErrors {
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
