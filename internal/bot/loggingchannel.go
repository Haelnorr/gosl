package bot

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	channelLogName string = "gosl-bot-log"
)

func (b *Bot) setupLogChannel(
	ctx context.Context,
) error {
	b.logger.Debug().Msg("Setting up log channel")
	channelID, err := b.ensureLogChannel(ctx)
	if err != nil {
		return errors.Wrap(err, "b.ensureLogChannel")
	}
	b.setLogChannel(channelID)
	b.logger.Info().Msg("Log channel setup complete")
	return nil
}

func (b *Bot) ensureLogChannel(
	ctx context.Context,
) (string, error) {
	b.logger.Debug().Msg("Ensuring log channel exists")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	defer tx.Commit()
	b.logger.Debug().Msg("Getting log channel ids")
	channelIDs, err := queryChannelsForPurpose(ctx, tx, channelLog)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	for _, channelID := range channelIDs {
		if exists := b.checkChannelExists(channelID); exists {
			b.logger.Debug().Msg("Log channel found")
			return channelIDs[0], nil
		} else {
			b.logger.Debug().Msg("Removing dead channel ID from database")
			removeChannelPurpose(ctx, tx, channelID, channelLog)
		}
	}

	b.logger.Debug().Msg("Creating new log channel")
	channel, err := b.session.GuildChannelCreate(
		b.guildID, channelLogName, discordgo.ChannelTypeGuildText)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "s.GuildChannelCreate")
	}

	b.logger.Debug().Msg("Adding new channel to database")
	err = addChannelPurpose(ctx, tx, channel.ID, channelLog)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "addChannelPurpose")
	}
	return channel.ID, nil
}
