package bot

import (
	"context"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	channelAdminName string = "gosl-bot-admin"
)

func (b *Bot) setupAdminChannel(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
) {
	defer wg.Done()
	b.logger.Debug().Msg("Setting up admin channel")
	channelID, err := b.ensureAdminChannel(ctx)
	if err != nil {
		errch <- errors.Wrap(err, "b.ensureAdminChannel")
		return
	}
	b.logger.Info().Str("channel_id", channelID).Msg("Admin channel is ready")

	err = b.updateAdminMessages(ctx, channelID)
	if err != nil {
		errch <- errors.Wrap(err, "b.updateAdminMessages")
		return
	}
	b.session.AddHandler(b.handleAdminChannelInteractions(ctx))
	b.logger.Info().Msg("Admin channel setup complete")
}

func (b *Bot) ensureAdminChannel(
	ctx context.Context,
) (string, error) {
	b.logger.Debug().Msg("Ensuring admin channel exists")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	defer tx.Commit()
	b.logger.Debug().Msg("Getting admin channel ids")
	channelIDs, err := queryChannelsForPurpose(ctx, tx, channelAdmin)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	for _, channelID := range channelIDs {
		if exists := b.checkChannelExists(channelID); exists {
			b.logger.Debug().Msg("Admin channel found")
			return channelIDs[0], nil
		} else {
			b.logger.Debug().Msg("Removing dead channel ID from database")
			removeChannelPurpose(ctx, tx, channelID, channelAdmin)
		}
	}

	b.logger.Debug().Msg("Creating new admin channel")
	channel, err := b.session.GuildChannelCreate(
		b.guildID, channelAdminName, discordgo.ChannelTypeGuildText)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "s.GuildChannelCreate")
	}

	b.logger.Debug().Msg("Adding new channel to database")
	err = addChannelPurpose(ctx, tx, channel.ID, channelAdmin)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "addChannelPurpose")
	}
	return channel.ID, nil
}

func (b *Bot) updateAdminMessages(
	ctx context.Context,
	channelID string,
) error {
	// Select log channel
	b.logger.Debug().Msg("Updating log channel select")
	err := b.updateChannelMessage(
		ctx,
		b.selectLogChannelContents,
		messageAdminSelectLogChannel,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectLogChannel)")
	}

	// Select admin roles
	b.logger.Debug().Msg("Updating admin roles select")
	err = b.updateChannelMessage(
		ctx,
		b.selectAdminRolesContents,
		messageAdminSelectAdminRoles,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectAdminRoles)")
	}

	// Select manager roles
	b.logger.Debug().Msg("Updating manager roles select")
	err = b.updateChannelMessage(
		ctx,
		b.selectManagerRolesContents,
		messageAdminSelectManagerRoles,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectManagerRoles)")
	}
	return nil
}
