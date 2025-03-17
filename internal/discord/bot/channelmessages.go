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

type Message struct {
	ID              string
	Label           string
	Purpose         uint16
	GetContents     MessageContentsFunc
	channel         *Channel
	bot             *Bot
	updateLock      bool
	updateCountdown int64
}

// Function for returning message contents for a complex message
// type MessageContents func() (string, *discordgo.MessageEmbed, []discordgo.MessageComponent)
type MessageContents struct {
	Embed      *discordgo.MessageEmbed
	Components []discordgo.MessageComponent
}

// Function that returns a MessageContents with the context and bot provided
type MessageContentsFunc func(ctx context.Context, b *Bot) (*MessageContents, error)

// Prepare the message by checking the database
func (m *Message) Setup(ctx context.Context, wg *sync.WaitGroup, errch chan error) {
	defer wg.Done()

	// DB setup
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := m.bot.Conn.RBegin(timeout)
	if err != nil {
		errch <- errors.Wrap(err, "conn.RBegin")
		return
	}
	defer tx.Rollback()

	// find an existing message in the database
	m.bot.Logger.Debug().Str("msg", m.Label).Msg("Finding existing message")
	messageID, channelID, err := models.GetMessageForPurpose(ctx, tx, m.Purpose)
	if err != nil {
		errch <- errors.Wrap(err, "models.GetMessageForPurpose")
		return
	}

	// Make sure the channel ID matches the expected channel and the message exists
	if channelID == m.channel.ID {
		if checkMessageExists(messageID, channelID, m.bot) {
			m.ID = messageID
		}
	}
	m.updateLock = true
	errch <- nil
}

func (m *Message) StartUpdate(addToQueue bool) bool {
	if m.updateLock && !addToQueue {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Message locked, cancelling")
		return false
	}
	if m.updateLock {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Message locked, queueing")
		timeToUpdate := time.Now().Add(time.Duration(500 * time.Millisecond)).Unix()
		m.updateCountdown = timeToUpdate
	} else {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Locking message")
		m.updateLock = true
	}
	return true
}
func (m *Message) EndUpdate() {
	m.bot.Logger.Debug().Str("msg", m.Label).Msg("Unlocking message")
	m.updateLock = false
}

// Updates (edits) the message with the discord API. Fails if message ID not found
func (m *Message) Update(ctx context.Context, errch chan error) {
	if !m.updateLock {
		errch <- errors.New("Message update was not started")
		return
	}
	// Get the time 500ms from now
	// Check if this message has an update request already queued
	if m.updateCountdown > time.Now().Unix() {
		// Abandon this update request
		return
	}
	// Initialise the timer
	timeToUpdate := time.Now().Add(time.Duration(500 * time.Millisecond)).Unix()
	m.updateCountdown = timeToUpdate
	defer func() {
		// On function exit, release the updateLock
		m.EndUpdate()
	}()
	// Wait for the update countdown to expire
	for m.updateCountdown < time.Now().Unix() {
		time.Sleep(50 * time.Millisecond)
	}
	// get the message contents
	contents, err := m.GetContents(ctx, m.bot)
	if err != nil {
		errch <- errors.Wrap(err, fmt.Sprintf("m.GetContents (%s)", m.Label))
		return
	}
	m.bot.Logger.Debug().Str("msg", m.Label).Msg("Updating message")

	// send the api request to edit the message
	m.bot.pool.queue()
	starttime := time.Now()
	_, err = m.bot.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		ID:         m.ID,
		Channel:    m.channel.ID,
		Embeds:     &[]*discordgo.MessageEmbed{contents.Embed},
		Components: &contents.Components,
	})
	m.bot.Logger.Debug().
		Dur("time_taken", time.Since(starttime)).
		Str("msg", m.Label).
		Msg("Message updated")
	if err != nil {
		errch <- errors.Wrap(err, "session.ChannelMessageEditComplex")
		return
	}
	errch <- nil
}

// Send the message with the discord API, updates the database with the message ID
func (m *Message) SendNew(ctx context.Context, errch chan error) {
	// get the message contents
	contents, err := m.GetContents(ctx, m.bot)
	if err != nil {
		errch <- errors.Wrap(err, fmt.Sprintf("m.GetContents (%s)", m.Label))
		return
	}
	m.bot.Logger.Debug().Str("msg", m.Label).Msg("Sending message")

	// send the api request to send the message
	m.bot.pool.queue()
	starttime := time.Now()
	message, err := m.bot.Session.ChannelMessageSendComplex(m.channel.ID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{contents.Embed},
		Components: contents.Components,
	})
	m.bot.Logger.Debug().
		Dur("time_taken", time.Since(starttime)).
		Str("msg", m.Label).
		Msg("Message sent")
	if err != nil {
		errch <- errors.Wrap(err, "session.ChannelMessageEditComplex")
		return
	}
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := m.bot.Conn.Begin(timeout)
	if err != nil {
		errch <- errors.Wrap(err, "conn.Begin")
		return
	}
	defer tx.Rollback()
	err = models.SetMessage(ctx, tx, message.ID, m.channel.ID, m.Purpose)
	if err != nil {
		errch <- errors.Wrap(err, "models.SetMessage")
		return
	}
	m.ID = message.ID
	tx.Commit()
	errch <- nil
}

// ===========================================================================
// PRIVATE FUNCTIONS
// ===========================================================================

// Check if a message exists with the discord API
func checkMessageExists(messageID, channelID string, b *Bot) bool {
	// b.apiQueue.queue()
	_, err := b.Session.ChannelMessage(channelID, messageID)
	return err == nil
}
