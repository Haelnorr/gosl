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
	ID               string
	Label            string
	Purpose          uint16
	GetContents      MessageContentsFunc
	channel          *Channel
	bot              *Bot
	updateLock       bool
	updateCountdown  int64
	updateInProgress bool
	updateQueue      uint16
	queueLock        sync.Mutex
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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := m.bot.Conn.RBegin(timeout, "Message.Setup(): "+m.Label)
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

// Signals that the message will be updated shortly, putting the message into
// a locked state so it cannot be modified until the update is completed.
// If addToQueue is set to false and an update is in progress, it will fail and
// return false. In all other cases it will return true (i.e. only needs to be
// checked if calling StartUpdate(false).
// If addToQueue is set to true and an update is already in progress, this will
// pause the update in progress until the update queue clears
func (m *Message) StartUpdate(addToQueue bool) bool {
	m.queueLock.Lock()
	defer m.queueLock.Unlock()
	if m.updateLock && !addToQueue {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Message locked, cancelling")
		return false
	}
	if m.updateLock {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Message locked, queueing")
		// timeToUpdate := time.Now().Add(time.Duration(750 * time.Millisecond)).Unix()
		// m.updateCountdown = timeToUpdate
		m.updateQueue += 1
		go func() {
			time.Sleep(10 * time.Second)
			if m.updateQueue != 0 {
				m.updateQueue = 0
				m.bot.Logger.Warn().Str("msg", m.Label).
					Msg("Message update queue never cleared. Message.Update() may never have been called. Manually clearing queue")
			}
		}()
	} else {
		m.bot.Logger.Debug().Str("msg", m.Label).Msg("Locking message")
		m.updateLock = true
	}
	return true
}
func (m *Message) endUpdate() {
	m.bot.Logger.Debug().Str("msg", m.Label).Msg("Unlocking message")
	m.updateLock = false
	m.updateInProgress = false
}

// Updates (edits) the message with the discord API. Fails if message ID not found.
// Message.StartUpdate() must be called first. If an update is already in progress
// and Message.StartUpdate(true) was called (adding update to queue), this is a
// basically a NOP that just removes the update from queue
func (m *Message) Update(ctx context.Context, errch chan error) {
	if !m.updateLock {
		errch <- errors.New(fmt.Sprintf("Message update was not started (%s)", m.Label))
		return
	}
	// Check if this message has an update request already queued
	if m.updateInProgress {
		// Abandon this update request, removing it from the queue
		m.updateQueue -= 1
		// push back the update timer in case more updates come in
		timeToUpdate := time.Now().Add(time.Duration(750 * time.Millisecond)).UnixMilli()
		m.updateCountdown = timeToUpdate
		return
	}
	m.updateInProgress = true
	defer func() {
		// On function exit, release the updateLock
		m.endUpdate()
	}()
	// Initialise the timer
	start := time.Now()
	timeToUpdate := start.Add(time.Duration(750 * time.Millisecond)).UnixMilli()
	m.updateCountdown = timeToUpdate
	// Wait for the update countdown to expire
	for m.updateCountdown > time.Now().UnixMilli() || m.updateQueue != 0 {
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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := m.bot.Conn.Begin(timeout, "Message.SendNew(): "+m.Label)
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
