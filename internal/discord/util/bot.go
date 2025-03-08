package util

import (
	"context"
	"gosl/pkg/db"
	"io/fs"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

// Contains the session objects for a created bot
type Bot struct {
	Session    *discordgo.Session
	Logchannel string
	Logger     *zerolog.Logger
	Files      *fs.FS
	Conn       *db.SafeConn
	GuildID    string
}

// Function for setting up a bot package concurrently
type SetupFunc func(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *Bot,
)
