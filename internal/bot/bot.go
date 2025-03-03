package bot

import (
	"context"
	"gosl/pkg/db"
	"io/fs"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Bot struct {
	session  *discordgo.Session
	logger   *zerolog.Logger
	commands []*command
	files    *fs.FS
	conn     *db.SafeConn
	guildID  string
}

func NewBot(
	token string,
	guildID string,
	conn *db.SafeConn,
	logger *zerolog.Logger,
	fs *fs.FS,
) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, errors.Wrap(err, "discordgo.New")
	}
	b := &Bot{session: session, logger: logger, files: fs, conn: conn, guildID: guildID}
	b.setupCommands()
	return b, nil
}

func (b *Bot) Start(ctx context.Context) error {
	err := b.session.Open()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	errch := make(chan error, 10)

	wg.Add(1)
	go b.setupAdminChannel(&wg, errch, ctx)
	wg.Add(1)
	go b.registerCommands(&wg, errch, ctx)

	go func() {
		wg.Wait()
		close(errch)
	}()

	hadErrors := false
	for err := range errch {
		if err != nil {
			b.logger.Error().Err(err).Msg("Error in bot startup")
			hadErrors = true
		}
	}
	if hadErrors {
		return errors.New("Error(s) during bot startup")
	}
	b.logger.Info().Msg("Bot startup complete!")
	return nil
}

func (b *Bot) Stop() error {
	err := b.session.Close()
	if err != nil {
		return err
	}
	return nil
}
