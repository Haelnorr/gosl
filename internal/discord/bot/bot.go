package bot

import (
	"context"
	"gosl/pkg/config"
	"gosl/pkg/db"
	"io/fs"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Contains the session objects for a created bot
type Bot struct {
	Session  *discordgo.Session
	Logger   *zerolog.Logger
	Files    *fs.FS
	Conn     *db.SafeConn
	Config   *config.Config
	Channels map[uint16]*Channel
}

// Create a new bot and start a session
func NewBot(
	l *zerolog.Logger,
	f *fs.FS,
	c *db.SafeConn,
	cfg *config.Config,
) (*Bot, error) {
	session, err := discordgo.New("Bot " + cfg.DiscordBotToken)
	if err != nil {
		return nil, errors.Wrap(err, "discordgo.New")
	}
	return &Bot{
		Session:  session,
		Logger:   l,
		Files:    f,
		Conn:     c,
		Config:   cfg,
		Channels: make(map[uint16]*Channel),
	}, nil
}

// Add a new channel to the bot. Fails if a channel with the same purpose has
// already been added
func (b *Bot) AddChannel(c *Channel) error {
	if _, exists := b.Channels[c.Purpose]; exists {
		return errors.New("Channel with that purpose already added.")
	}
	c.bot = b
	b.Channels[c.Purpose] = c
	return nil
}

// Function for setting up a bot package
type SetupFunc func(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *Bot,
)
