package logchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	logChannelName string = "gosl-bot-log"
)

// Setup the logging channel
func Setup(
	ctx context.Context,
	b *util.Bot,
) error {
	b.Logger.Debug().Msg("Setting up log channel")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return errors.Wrap(err, "conn.Begin")
	}
	defer tx.Rollback()
	channelID, err := findExistingLogChannel(ctx, tx, b)
	if err != nil {
		return errors.Wrap(err, "findExistingLogChannel")
	}
	if channelID == "" {
		channelID, err = createNewLogChannel(ctx, tx, b)
		if err != nil {
			return errors.Wrap(err, "createNewLogChannel")
		}
		if !channels.CheckExists(channelID, b.Session) {
			return errors.New("Unknown error occurred setting up log channel")
		}
	}
	tx.Commit()
	b.Logchannel = channelID
	b.Logger.Info().Msg("Log channel setup complete")
	return nil
}

func findExistingLogChannel(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
) (string, error) {
	b.Logger.Debug().Msg("Getting log channel ids")
	channelIDs, err := channels.GetChannels(ctx, tx, channels.PurposeLog)
	if err != nil {
		return "", errors.Wrap(err, "channels.GetChannels")
	}
	var logChannelID string
	for _, channelID := range channelIDs {
		if exists := channels.CheckExists(channelID, b.Session); exists {
			b.Logger.Debug().Msg("Log channel found")
			logChannelID = channelID
		} else {
			b.Logger.Debug().Msg("Removing dead channel ID from database")
			channels.RemovePurpose(ctx, tx, channelID, channels.PurposeLog)
		}
	}
	return logChannelID, nil
}

func createNewLogChannel(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
) (string, error) {
	b.Logger.Debug().Msg("Creating new log channel")
	channel, err := b.Session.GuildChannelCreate(
		b.GuildID, logChannelName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", errors.Wrap(err, "s.GuildChannelCreate")
	}

	b.Logger.Debug().Msg("Adding new channel to database")
	err = channels.AddPurpose(ctx, tx, channel.ID, channels.PurposeLog)
	if err != nil {
		return "", errors.Wrap(err, "channels.AddPurpose")
	}
	return channel.ID, nil
}
