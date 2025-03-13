package bot

import (
	"context"
	"fmt"
	"gosl/internal/models"
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
	apiQueue sync.Mutex
}

// Setups up the channel in discord, finding an existing channel or creating
// a new one if no existing channel can be found
func (c *Channel) Setup(ctx context.Context) error {
	if c.bot == nil {
		return errors.New(fmt.Sprintf("Channel not registered to bot (%s)", c.Label))
	}
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Setting up channel")
	// Find an existing channel
	channelID, err := c.findExisting(ctx)
	if err != nil {
		return errors.Wrap(err, "FindExisting")
	}
	if channelID == "" {
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

// TODO: these dont appear to help, leave them in for now but remove if unnecessary
// Lock the api request queue for this channel
func (c *Channel) LockQueue() {
	// c.apiQueue.Lock()
}

// Unlock the api request queue for this channel
func (c *Channel) ReleaseQueue() {
	// NOTE: Tweak this delay for best results
	go func() {
		// time.Sleep(500 * time.Millisecond)
		// c.apiQueue.Unlock()
	}()
}

// Updates the channel ID of the channel and saves it in the database
func (c *Channel) UpdateTarget(ctx context.Context, newID string) error {
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := c.bot.Conn.Begin(timeout)
	if err != nil {
		return errors.Wrap(err, "conn.Begin")
	}
	defer tx.Rollback()
	err = models.SetChannel(ctx, tx, newID, c.Purpose)
	if err != nil {
		return errors.Wrap(err, "models.SetChannel")
	}
	c.ID = newID
	// TODO: delete all messages from discord (warn gracefully if message not found)
	// TODO: delete all messages from database
	// TODO: set all message IDs to ""
	// TODO: resend all messages to new channel
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

// Prepares and updates all the messages in the channel. Should only be run on startup
func (c *Channel) SetupMessages(ctx context.Context, errch chan error) {
	var wg sync.WaitGroup
	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Setting up messages")
	for _, message := range c.Messages {
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
	msg, embed, components := contents()
	_, err := c.bot.Session.ChannelMessageSendComplex(c.ID, &discordgo.MessageSend{
		Content:    msg,
		Embeds:     []*discordgo.MessageEmbed{embed},
		Components: components,
	})
	if err != nil {
		return errors.Wrap(err, "session.ChannelMessageSendComplex")
	}
	return nil
}

// ===========================================================================
// PRIVATE FUNCTIONS
// ===========================================================================

// Find an existing channel for the provided purpose and return the channel id
func (c *Channel) findExisting(ctx context.Context) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := c.bot.Conn.RBegin(timeout)
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
		if exists := checkChannelExists(channelID, c.bot.Session); exists {
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
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := c.bot.Conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	defer tx.Rollback()

	c.bot.Logger.Debug().Str("channel", c.Label).Msg("Creating new channel")
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
func checkChannelExists(channelID string, s *discordgo.Session) bool {
	if channelID == "" {
		return false
	}
	_, err := s.Channel(channelID)
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
	tx, err := b.Conn.Begin(timeout)
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
