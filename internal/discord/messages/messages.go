package messages

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	AdminSelectLogChannel   uint16 = 1 // Admin channel: select log channel component
	AdminSelectAdminRoles   uint16 = 2 // Admin channel: select admin roles component
	AdminSelectManagerRoles uint16 = 3 // Admin channel: select manager roles component

	ManagerSelectSeason uint16 = 4 // Manager channel: select season component
	ManagerCreateSeason uint16 = 5 // Manager channel: create season component
	ManagerActiveSeason uint16 = 6 // Manager channel: active season component
)

// Set the provided message as the message used for the provided purpose
// Only one message can be used for a given purpose at a time
// Setting a new message will overwrite the previous one
func SetMessage(
	ctx context.Context,
	tx *db.SafeWTX,
	messageID string,
	channelID string,
	purpose uint16,
) error {
	query := `
INSERT INTO config_messages (message_id, channel_id, purpose) 
VALUES (?, ?, ?) 
ON CONFLICT(purpose) DO UPDATE
SET message_id = excluded.message_id,
    channel_id = excluded.channel_id;
`
	_, err := tx.Exec(ctx, query, messageID, channelID, purpose)
	return err
}

// Remove the provided message from the database
func RemoveMessage(
	ctx context.Context,
	tx *db.SafeWTX,
	messageID string,
	channelID string,
	purpose uint16,
) error {
	query := `
DELETE FROM config_messages WHERE message_id = ? AND channel_id = ? AND purpose = ?;
`
	_, err := tx.Exec(ctx, query, messageID, channelID, purpose)
	return err
}

// Get the message that has been set for the provided purpose
// Returns messageID, channelID, err
func GetMessageForPurpose(
	ctx context.Context,
	tx db.SafeTX,
	purpose uint16,
) (string, string, error) {
	query := `
SELECT message_id, channel_id FROM config_messages WHERE purpose = ?;
`
	var messageID string
	var channelID string
	row, err := tx.QueryRow(ctx, query, purpose)
	if err != nil {
		return "", "", errors.Wrap(err, "tx.QueryRow")
	}
	err = row.Scan(&messageID, &channelID)
	if err == sql.ErrNoRows {
		return "", "", nil
	}
	if err != nil {
		return "", "", errors.Wrap(err, "row.Scan")
	}

	return messageID, channelID, nil
}

// Check if a message exists with the discord API
func CheckMessageExists(messageID, channelID string, s *discordgo.Session) bool {
	_, err := s.ChannelMessage(channelID, messageID)
	return err == nil
}

type ChannelMessage struct {
	Label        string
	Purpose      uint16
	Channel      uint16
	ContentsFunc util.MessageContentsFunc
}

// For the given channel, iterate through the provided messages and update them
// in discord. If existing messages can't be found they will be created, otherwise
// existing messages will be edited
func UpdateChannelMessages(
	ctx context.Context,
	b *util.Bot,
	messages []*ChannelMessage,
) []error {
	var wg sync.WaitGroup
	var qwg sync.WaitGroup
	errch := make(chan error)
	orderCh := make(chan int, 1)
	orderCh <- 0
	for i, msg := range messages {
		b.Logger.Debug().Str("message", msg.Label).Msg("Updating message")
		wg.Add(1)
		qwg.Add(1)
		// to ensure new messages are created in the correct order, use the index
		// of the messages array to setup a queue. any work that can be done
		// concurrently will run and if a message needs to be created, it will
		// queue up in the order provided by the messages array
		go func(i int, msg *ChannelMessage) {
			defer wg.Done()
			err := UpdateChannelMessageConcurrently(&qwg, i, orderCh, ctx, b, msg)
			if err != nil {
				errch <- errors.Wrap(err, "UpdateChannelMessageConcurrently")
			}
		}(i, msg)
	}
	go func() {
		wg.Wait()
		close(errch)
	}()

	var msgerrors []error
	for err := range errch {
		if err != nil {
			msgerrors = append(msgerrors, err)
		}
	}
	return msgerrors
}

// Update the provided message using the provided contents function
func UpdateChannelMessageConcurrently(
	wg *sync.WaitGroup,
	i int,
	orderCh chan int,
	ctx context.Context,
	b *util.Bot,
	message *ChannelMessage,
) error {
	// update the message
	err := addOrEditChannelMessage(wg, i, orderCh, ctx, b, message)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("AddOrEditChannelMessage (%v)", message.Label))
	}
	return nil
}

// Update the provided message using the provided contents function
func UpdateChannelMessage(
	ctx context.Context,
	b *util.Bot,
	message *ChannelMessage,
) error {
	// faking a message queue and putting us first in line
	i := 0
	orderCh := make(chan int, 1)
	orderCh <- 0
	var wg sync.WaitGroup
	// update the message
	wg.Add(1)
	err := addOrEditChannelMessage(&wg, i, orderCh, ctx, b, message)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("AddOrEditChannelMessage (%s)", message.Label))
	}
	go func() {
		wg.Wait()
		close(orderCh)
	}()
	return nil
}

// Edit the message for the provided purpose; if it doesn't exist, create a new one
func addOrEditChannelMessage(
	wg *sync.WaitGroup,
	i int,
	orderCh chan int,
	ctx context.Context,
	b *util.Bot,
	message *ChannelMessage,
) error {
	b.Logger.Debug().Str("message", message.Label).Msg("Updating message")
	// get the message components
	contents, err := message.ContentsFunc(ctx, b)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("message.ContentsFunc (%s)", message.Label))
	}
	// attempt to edit an existing message
	err = editChannelMessage(ctx, b, message, contents)
	if !(err != nil && err.Error() == "No message found") {
		// either message was found or critical error occured:
		// spin off a routine that waits in the queue until it's next in line
		// and then signals it's finished. if we don't do this then the queue
		// will get updated out of order and messages that are required to
		// wait will hang indefinitely
		go func() {
			for {
				current := <-orderCh
				if current == i {
					orderCh <- i + 1 // signal to next in queue we are done
					wg.Done()
					break
				}
				orderCh <- current // not our turn yet
				// random delay
				time.Sleep(time.Duration(30+rand.Intn(31)) * time.Millisecond)
			}
		}()
		if err == nil {
			// message edited successfully
			return nil
		}
		if err.Error() != "No message found" {
			// critical error
			return errors.Wrap(err, "editChannelMessage")
		}
	}
	b.Logger.Debug().Str("message", message.Label).
		Msg("No message found, creating new message")
	// queue until it's this functions turn to create a new message
	for {
		current := <-orderCh // pop the current index from the channel
		if current == i {
			break // its our turn
		}
		orderCh <- current // not our turn yet, put the current index back
		// random delay to improve chance correct function gets triggered next
		// to be honest not sure if this actually helps but seems like a good idea lol
		time.Sleep(time.Duration(30+rand.Intn(31)) * time.Millisecond)
	}
	defer func() {
		orderCh <- i + 1
		wg.Done()
	}() // signal to next in queue we are done
	// create a new channel message
	err = addChannelMessage(ctx, b, message, contents)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("addChannelMessage (%s)", message.Label))
	}
	return nil
}

// Attempt to create a new message with the discord API and add it to the database
func addChannelMessage(
	ctx context.Context,
	b *util.Bot,
	message *ChannelMessage,
	contents util.MessageContents,
) error {
	// get the channel ID for the message
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return errors.Wrap(err, "b.RConn.Begin")
	}
	channelID, err := channels.GetChannel(ctx, tx, message.Channel)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "channels.GetChannel")
	}
	tx.Commit()
	// send a request to the discord API to create a message
	messageID, err := CreateComplexMessage(contents, channelID, b.Session)
	if err != nil {
		return errors.Wrap(err, "CreateComplex")
	}
	// add the message ID to the database for future reference
	b.Logger.Debug().Str("message", message.Label).Msg("Adding message to database")
	timeout, cancel = context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	wtx, err := b.Conn.Begin(timeout)
	if err != nil {
		return errors.Wrap(err, "b.Conn.Begin")
	}
	err = SetMessage(ctx, wtx, messageID, channelID, message.Purpose)
	if err != nil {
		wtx.Rollback()
		return errors.Wrap(err, "SetMessage")
	}
	wtx.Commit()
	b.Logger.Debug().Str("message", message.Label).Msg("Added message to database")
	return nil
}

// Attempt to find a message in the database and edit it with the discord API
func editChannelMessage(
	ctx context.Context,
	b *util.Bot,
	message *ChannelMessage,
	contents util.MessageContents,
) error {
	// find an existing message in the database
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Str("message", message.Label).Msg("Finding existing message")
	messageID, channelID, err := GetMessageForPurpose(
		ctx, tx, message.Purpose)
	if err != nil {
		return errors.Wrap(err, "getMessageForPurpose")
	}
	if messageID != "" && channelID != "" {
		// message found, attempt to edit the message with the discord API
		b.Logger.Debug().Str("message", message.Label).Msg("Existing message in DB")
		err = EditComplexMessage(contents, messageID, channelID, b.Session)
		if err != nil {
			if strings.Contains(err.Error(), "HTTP 404 Not Found") {
				// message does not exist
				return errors.New("No message found")
			}
			// unexpected error
			return errors.Wrap(err, "b.editStaticMessage")
		}
		// message edited successfully
		return nil
	}
	// no message was found
	return errors.New("No message found")
}
