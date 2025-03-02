package bot

import (
	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Bot struct {
	session  *discordgo.Session
	logger   *zerolog.Logger
	commands []*command
}

func NewBot(token string, logger *zerolog.Logger) (*Bot, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, errors.Wrap(err, "discordgo.New")
	}
	b := &Bot{session: session, logger: logger}
	b.setupCommands()
	return b, nil
}

func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return err
	}
	err = b.clearCommands()
	if err != nil {
		b.logger.Error().Err(err).Msg("Failed to clear old commands")
	}
	err = b.registerCommands()
	if err != nil {
		b.logger.Error().Err(err).Msg("Failed to register commands")
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
