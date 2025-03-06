package channels

import (
	"context"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Setup a channel for the given purpose
func Setup(
	ctx context.Context,
	b *util.Bot,
	purpose uint16,
	name string,
) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	channelID, err := FindExisting(ctx, tx, b, purpose)
	if err != nil {
		return "", errors.Wrap(err, "FindExisting")
	}
	if channelID == "" {
		channelID, err = CreateNew(ctx, tx, b, purpose, name)
		if err != nil {
			return "", errors.Wrap(err, "CreateNew")
		}
		if !CheckExists(channelID, b.Session) {
			return "", errors.New("Unknown error occurred setting up the channel")
		}
	}
	tx.Commit()
	return channelID, nil
}

// Find an existing channel for the provided purpose and return the channel id
func FindExisting(
	ctx context.Context,
	tx *db.SafeWTX,
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
	for _, channelID := range channelIDs {
		if exists := CheckExists(channelID, b.Session); exists {
			b.Logger.Debug().Str("channel", purposeName).Msg("Channel found")
			selectedChannelID = channelID
		} else {
			b.Logger.Debug().Msg("Removing dead channel ID from database")
			RemovePurpose(ctx, tx, channelID, purpose)
		}
	}
	return selectedChannelID, nil
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
		b.GuildID, name, discordgo.ChannelTypeGuildText)
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
