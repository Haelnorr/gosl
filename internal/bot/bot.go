package bot

import (
	"context"
	"gosl/pkg/db"
	"io/fs"

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

	err = b.setupAdminChannel(ctx)
	if err != nil {
		return errors.Wrap(err, "b.setupAdminChannel")
	}
	err = b.registerCommands(ctx)
	if err != nil {
		return errors.Wrap(err, "b.registerCommands")
	}
	return nil
}

func (b *Bot) Stop() error {
	err := b.session.Close()
	if err != nil {
		return err
	}
	return nil
}
