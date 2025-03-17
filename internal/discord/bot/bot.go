package bot

import (
	"context"
	"fmt"
	"gosl/pkg/config"
	"gosl/pkg/db"
	"io/fs"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Contains the session objects for a created bot
type Bot struct {
	Session         *discordgo.Session
	Logger          *zerolog.Logger
	Files           *fs.FS
	Conn            *db.SafeConn
	Config          *config.Config
	Channels        map[uint16]*Channel
	DirectMessages  map[string]*DirectMessage
	DynamicMessages map[string]*DynamicMessage
	pool            *requestPool
}

// Function for setting up a bot package
type SetupFunc func(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *Bot,
)

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
	bot := &Bot{
		Session:         session,
		Logger:          l,
		Files:           f,
		Conn:            c,
		Config:          cfg,
		Channels:        make(map[uint16]*Channel),
		DirectMessages:  make(map[string]*DirectMessage),
		DynamicMessages: make(map[string]*DynamicMessage),
		pool:            newRequestPool(),
	}
	return bot, nil
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

func (b *Bot) GetMessage(
	channelPurpose uint16,
	messagePurpose uint16,
) (*Message, error) {
	channel, exists := b.Channels[channelPurpose]
	if !exists {
		return nil, errors.New("Channel not found")
	}
	msg, exists := channel.Messages[messagePurpose]
	if !exists {
		return nil, errors.New("Message not found")
	}
	return msg, nil
}

func (b *Bot) addDirectMessage(dm *DirectMessage) error {
	_, exists := b.DirectMessages[dm.ID]
	if exists {
		return nil
	}
	b.DirectMessages[dm.ID] = dm
	return nil
}

func (b *Bot) removeDirectMessage(messageID string) error {
	_, exists := b.DirectMessages[messageID]
	if !exists {
		return nil
	}
	delete(b.DirectMessages, messageID)
	return nil
}

func (b *Bot) SendDirectMessage(title, message string, userID string) error {
	channel, err := b.Session.UserChannelCreate(userID)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.UserChannelCreate (%s)", userID))
	}
	_, err = b.Session.ChannelMessageSendComplex(channel.ID, &discordgo.MessageSend{
		Embed: &discordgo.MessageEmbed{
			Title:       title,
			Description: message,
		},
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("session.ChannelMessageSendComplex (%s)", userID))
	}
	return nil
}

func (b *Bot) GetDirectMessage(
	messageID string,
	userID string,
	label string,
	expiry time.Duration,
	deleteAfter bool,
) (*DirectMessage, error) {
	msg, exists := b.DirectMessages[messageID]
	if !exists {
		channel, err := b.Session.UserChannelCreate(userID)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("session.UserChannelCreate (%s, %s)", label, userID))
		}
		exists = checkMessageExists(messageID, channel.ID, b)
		if exists {
			msg, err = reAddDirectMessage(label, messageID, userID, channel, expiry, deleteAfter, b)
			if err != nil {
				return nil, errors.Wrap(err, "reAddDirectMessage")
			}
			return msg, nil
		}
		return nil, errors.New("Message not found")
	}
	return msg, nil
}

func (b *Bot) addDynamicMessage(dy *DynamicMessage) error {
	_, exists := b.DynamicMessages[dy.ID]
	if exists {
		return nil
	}
	b.DynamicMessages[dy.ID] = dy
	return nil
}

func (b *Bot) removeDynamicMessage(messageID string) error {
	_, exists := b.DynamicMessages[messageID]
	if !exists {
		return nil
	}
	delete(b.DynamicMessages, messageID)
	return nil
}

func (b *Bot) GetDynamicMessage(label, messageID, channelID string,
) (*DynamicMessage, error) {
	msg, exists := b.DynamicMessages[messageID]
	if !exists {
		exists = checkMessageExists(messageID, channelID, b)
		if exists {
			msg = reAddDynamicMessage(label, messageID, channelID, b)
			return msg, nil
		}
		return nil, errors.New("Message not found")
	}
	return msg, nil
}
