package bot

import (
	"context"
	"fmt"
	"gosl/internal/models"
	"gosl/pkg/db"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Represents a text channel in discord
type Channel struct {
	ID       string
	Purpose  uint16
	Name     string
	Label    string
	Messages map[uint16]*Message
	Handler  Handler
	bot      *Bot
}

// Setups up the channel in discord, finding an existing channel or creating
// a new one if no existing channel can be found
func (c *Channel) Setup(ctx context.Context, forceCreate bool) error {
	if c.bot == nil {
		return errors.New(fmt.Sprintf("Channel not registered to bot (%s)", c.Label))
	}
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Setting up channel")
	// Find an existing channel
	channelID, err := c.findExisting(ctx)
	if err != nil {
		return errors.Wrap(err, "FindExisting")
	}
	if channelID == "" && forceCreate {
		// Setup a new channel
		channelID, err = c.createNew(ctx)
		if err != nil {
			return errors.Wrap(err, "CreateNew")
		}
	}

	c.ID = channelID
	c.Messages = make(map[uint16]*Message)
	c.bot.Logger.Debug().Str("channel_id", channelID).Str("channel", c.Label).
		Msg("Channel ready")
	return nil
}

// Updates the channel ID of the channel and saves it in the database
func (c *Channel) UpdateTarget(ctx context.Context, tx *db.SafeWTX, newID string) error {
	err := models.SetChannel(ctx, tx, newID, c.Purpose)
	if err != nil {
		return errors.Wrap(err, "models.SetChannel")
	}
	for _, message := range c.Messages {
		c.DeleteMessage(message)
		err = models.RemoveMessage(ctx, tx, message.ID, c.ID, message.Purpose)
		if err != nil {
			return errors.Wrap(err, "models.RemoveMessage")
		}
		message.ID = ""
	}
	c.ID = newID
	go func() {
		errch := make(chan error)
		var wg sync.WaitGroup
		wg.Add(1)
		go c.SetupMessages(ctx, &wg, errch)
		go func() {
			wg.Wait()
			close(errch)
		}()
		for err := range errch {
			if err != nil {
				c.bot.Logger.Error().Err(err).Msg("Failed sending message")
			}
		}
	}()
	return nil
}

// Register a message to the channel. Fails if a message with the given purpose
// has already been registered. Requires channel to be registered to the bot
func (c *Channel) RegisterMessage(m *Message) error {
	if c.bot == nil {
		return errors.New(fmt.Sprintf("Channel not registered to bot (%s)", c.Label))
	}
	if c.Messages == nil {
		return errors.New(fmt.Sprintf("Channel setup not called (%s)", c.Label))
	}
	if _, exists := c.Messages[m.Purpose]; exists {
		return errors.New("Message with that purpose already registered")
	}
	m.channel = c
	m.bot = c.bot
	c.Messages[m.Purpose] = m
	return nil
}

// Prepares and updates all the messages in the channel.
// func (c *Channel) SetupMessages(ctx context.Context, errch chan error) {
func (c *Channel) SetupMessages(ctx context.Context, swg *sync.WaitGroup, errch chan error) {
	defer swg.Done()
	if c.ID == "" {
		c.bot.Logger.Debug().Str("channel", c.Label).Msg("Channel not set, skipping message updates")
		return
	}
	var wg sync.WaitGroup
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Setting up messages")
	for _, message := range c.Messages {
		c.bot.Logger.Debug().Str("msg", message.Label).Msg("Setting up message")
		wg.Add(1)
		go func() {
			defer wg.Done()
			var wg sync.WaitGroup
			wg.Add(1)
			message.Setup(ctx, &wg, errch)
			wg.Wait()
			if message.ID != "" {
				message.Update(ctx, errch)
			} else {
				message.SendNew(ctx, errch)
			}
		}()
	}
	wg.Wait()
	c.bot.Session.AddHandler(c.Handler)
}

// Sends a message to the channel
func (c *Channel) SendMessage(
	contents MessageContents,
) error {
	c.bot.pool.queue()
	_, err := c.bot.Session.ChannelMessageSendComplex(c.ID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{contents.Embed},
		Components: contents.Components,
	})
	if err != nil {
		return errors.Wrap(err, "session.ChannelMessageSendComplex")
	}
	return nil
}

// Sends a message with a file to the channel
func (c *Channel) SendFile(message, filename string, file io.Reader) (*discordgo.Message, error) {
	c.bot.pool.queue()
	msg, err := c.bot.Session.ChannelFileSendWithMessage(c.ID, message, filename, file)
	if err != nil {
		return nil, errors.Wrap(err, "session.ChannelFileSendWithMessage")
	}
	return msg, nil
}

func (c *Channel) DeleteMessage(m *Message) {
	err := c.bot.Session.ChannelMessageDelete(c.ID, m.ID)
	if err != nil {
		if !strings.Contains(err.Error(), "HTTP 404 Not Found") {
			c.bot.Logger.Warn().Err(err).Str("msg", m.Label).
				Msg("Failed to delete message in discord")
		}
	}
}

// ===========================================================================
// PRIVATE FUNCTIONS
// ===========================================================================

// Find an existing channel for the provided purpose and return the channel id
func (c *Channel) findExisting(ctx context.Context) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := c.bot.Conn.RBegin(timeout, "Channel.findExisting: "+c.Label)
	if err != nil {
		return "", errors.Wrap(err, "conn.RBegin")
	}
	defer tx.Rollback()
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Getting channel ids")
	channelIDs, err := models.GetChannels(ctx, tx, c.Purpose)
	if err != nil {
		return "", errors.Wrap(err, "models.GetChannels")
	}
	var selectedChannelID string
	deadChannels := []string{}
	for _, channelID := range channelIDs {
		if exists := checkChannelExists(channelID, c.bot); exists {
			c.bot.Logger.Debug().Str("channel", c.Label).Msg("Channel found")
			selectedChannelID = channelID
		} else {
			deadChannels = append(deadChannels, channelID)
		}
	}
	if len(deadChannels) > 0 {
		go cleanupDeadChannels(ctx, c.bot, deadChannels, c.Purpose)
	}
	return selectedChannelID, nil
}

func (c *Channel) createNew(ctx context.Context) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := c.bot.Conn.Begin(timeout, "Channel.createNew(): "+c.Label)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	defer tx.Rollback()

	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Creating new channel")
	c.bot.pool.queue()
	channel, err := c.bot.Session.GuildChannelCreate(
		c.bot.Config.DiscordGuildID, c.Name, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", errors.Wrap(err, "session.GuildChannelCreate")
	}
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Adding new channel to database")
	err = models.AddChannel(ctx, tx, channel.ID, c.Purpose)
	if err != nil {
		return "", errors.Wrap(err, "models.AddPurpose")
	}
	tx.Commit()
	return channel.ID, nil
}

// Check with the discord API if the channel exists
func checkChannelExists(channelID string, b *Bot) bool {
	if channelID == "" {
		return false
	}
	// b.apiQueue.queue()
	_, err := b.Session.Channel(channelID)
	return err == nil
}

// Removes the provided channel IDs from the database
func cleanupDeadChannels(
	ctx context.Context,
	b *Bot,
	channelIDs []string,
	purpose uint16,
) {
	b.Logger.Debug().Msg("Removing dead channels from database")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout, "cleanupDeadChannels")
	if err != nil {
		b.Logger.Warn().Err(err).Msg("Failed to cleanup dead channels")
		return
	}
	for _, channelID := range channelIDs {
		b.Logger.Debug().Msg("Removing dead channel ID from database")
		models.RemoveChannel(ctx, tx, channelID, purpose)
	}
	tx.Commit()
}
