package channels

import (
	"context"
	"fmt"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

type UpdateMsgFunc func(ctx context.Context, b *util.Bot) []error

// Setup a channel for the given purpose
func Setup(
	errch chan error,
	ctx context.Context,
	b *util.Bot,
	purpose uint16,
	name string,
	updateFunc UpdateMsgFunc,
	interactionHandler util.Handler,
) {
	// Set the channel up
	channelName := PurposeName(purpose)
	b.Logger.Debug().Str("channel", channelName).Msg("Setting up channel")
	channelID, err := CreateOrFindChannel(ctx, b, purpose, name)
	if err != nil {
		errch <- errors.Wrap(err, fmt.Sprintf("createOrFindChannel (%s)", channelName))
		return
	}
	b.Logger.Debug().
		Str("channel_id", channelID).
		Str("channel", channelName).
		Msg("Channel ready")

		// Update the channel messages
	msgerrors := updateFunc(ctx, b)
	for _, err = range msgerrors {
		if err != nil {
			errch <- errors.Wrap(err, fmt.Sprintf("updateFunc (%s)", channelName))
		}
	}
	if len(msgerrors) > 0 {
		return
	}
	// Handle the interactions
	b.Session.AddHandler(interactionHandler)
	b.Logger.Info().Str("channel", channelName).Msg("Channel setup complete")
}

// Create or find a channel for the given purpose
func CreateOrFindChannel(
	ctx context.Context,
	b *util.Bot,
	purpose uint16,
	name string,
) (string, error) {
	// Find an existing channel
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "b.conn.RBegin")
	}
	defer tx.Rollback()
	channelID, err := FindExisting(ctx, tx, b, purpose)
	if err != nil {
		return "", errors.Wrap(err, "FindExisting")
	}
	tx.Commit()

	// Channel doesnt exist, create a new one
	if channelID == "" {
		timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout)
		if err != nil {
			return "", errors.Wrap(err, "b.conn.RBegin")
		}
		defer tx.Rollback()
		channelID, err = CreateNew(ctx, tx, b, purpose, name)
		if err != nil {
			return "", errors.Wrap(err, "CreateNew")
		}
		if !CheckExists(channelID, b.Session) {
			return "", errors.New("Unknown error occurred setting up the channel")
		}
		tx.Commit()
	}
	return channelID, nil
}

// Find an existing channel for the provided purpose and return the channel id
func FindExisting(
	ctx context.Context,
	tx *db.SafeRTX,
	b *util.Bot,
	purpose uint16,
) (string, error) {
	purposeName := PurposeName(purpose)
	b.Logger.Debug().Str("channel", purposeName).Msg("Getting channel ids")
	channelIDs, err := GetChannels(ctx, tx, purpose)
	if err != nil {
		return "", errors.Wrap(err, "GetChannels")
	}
	var selectedChannelID string
	deadChannels := []string{}
	for _, channelID := range channelIDs {
		if exists := CheckExists(channelID, b.Session); exists {
			b.Logger.Debug().Str("channel", purposeName).Msg("Channel found")
			selectedChannelID = channelID
		} else {
			deadChannels = append(deadChannels, channelID)
		}
	}
	if len(deadChannels) > 0 {
		go CleanupDeadChannels(ctx, b, deadChannels, purpose)
	}
	return selectedChannelID, nil
}

func CleanupDeadChannels(
	ctx context.Context,
	b *util.Bot,
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
		RemovePurpose(ctx, tx, channelID, purpose)
	}
	tx.Commit()
}

// Create a new channel for the provided purpose
func CreateNew(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	purpose uint16,
	name string,
) (string, error) {
	purposeName := PurposeName(purpose)
	b.Logger.Debug().Str("channel", purposeName).Msg("Creating new channel")
	channel, err := b.Session.GuildChannelCreate(
		b.Config.DiscordGuildID, name, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", errors.Wrap(err, "b.Session.GuildChannelCreate")
	}

	b.Logger.Debug().Str("channel", purposeName).Msg("Adding new channel to database")
	err = AddPurpose(ctx, tx, channel.ID, purpose)
	if err != nil {
		return "", errors.Wrap(err, "AddPurpose")
	}
	return channel.ID, nil
}
