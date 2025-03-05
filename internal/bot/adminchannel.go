package bot

import (
	"context"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	channelAdminName string = "gosl-bot-admin"
)

func (b *Bot) handleAdminChannelInteractions(ctx context.Context) handler {
	b.logger.Debug().Msg("Adding handler for admin channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			channelID, err := b.getChannel(ctx, channelAdmin)
			if err != nil {
				b.logger.Error().Err(err).Msg("failed to get a channel id for the admin channel")
				panic("unable to get admin channel ID")
			}
			if i.Message.ChannelID != channelID {
				return
			}
			b.logger.Debug().Msg("Handling admin channel interaction")
			timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			tx, err := b.conn.Begin(timeout)
			if err != nil {
				msg := "Failed to start database transaction"
				b.logger.Error().Err(err).Msg(msg)
				errorResponse("Database error occured", &msg, b.files, s, i)
				return
			}
			defer tx.Rollback()
			// TODO: allow only if user is admin (discord override) or has one
			// of the set admin roles

			// Handle select menu interactions
			switch i.MessageComponentData().CustomID {
			case "log_channel_select":
				b.logger.Debug().Msg("Handling log channel select interaction")
				err = handleSelectLogChannelInteraction(ctx, tx, b.logger, s, i)
				if err != nil {
					msg := "Failed to update log channel"
					smsg := err.Error()
					b.logger.Error().Err(err).Msg(msg)
					errorResponse(msg, &smsg, b.files, s, i)
				}
				tx.Commit()
			case "admin_role_select":
				// TODO: update the roles in the db
				b.logger.Debug().Msg("Handling admin roles select interaction")
				_ = i.MessageComponentData().Values
				emphemeralMessage("updated admin roles", b.logger, s, i)
			case "manager_role_select":
				// TODO: update the roles in the db
				b.logger.Debug().Msg("Handling manager roles select interaction")
				_ = i.MessageComponentData().Values
				emphemeralMessage("updated admin roles", b.logger, s, i)
			}
		}
	}
}

func (b *Bot) setupAdminChannel(ctx context.Context) error {
	b.logger.Debug().Msg("Setting up admin channel")
	channelID, err := b.ensureAdminChannel(ctx)
	if err != nil {
		return err
	}
	b.logger.Info().Str("channel_id", channelID).Msg("Admin channel is ready")

	err = b.updateAdminMessages(ctx, channelID)
	if err != nil {
		return errors.Wrap(err, "b.updateAdminMessages")
	}

	b.session.AddHandler(b.handleAdminChannelInteractions(ctx))
	return nil
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
	// Update the messages concurrently to avoid delays from database queries
	errch := make(chan (error))
	defer close(errch)
	// Select log channel
	b.logger.Debug().Msg("Updating log channel select")
	go b.updateChannelMessage(
		errch,
		ctx,
		b.selectLogChannelContents,
		messageAdminSelectLogChannel,
		channelID,
	)

	// Select admin roles
	b.logger.Debug().Msg("Updating admin roles select")
	go b.updateChannelMessage(
		errch,
		ctx,
		b.selectAdminRolesContents,
		messageAdminSelectAdminRoles,
		channelID,
	)

	// Select manager roles
	b.logger.Debug().Msg("Updating manager roles select")
	go b.updateChannelMessage(
		errch,
		ctx,
		b.selectManagerRolesContents,
		messageAdminSelectManagerRoles,
		channelID,
	)

	// Handle the errors
	var errorSlice []error
	for err := range errch {
		if err != nil {
			b.logger.Error().Err(err).Msg("Error updating admin channel message")
			errorSlice = append(errorSlice, err)
		}
	}
	if len(errorSlice) > 0 {
		return errors.New("Multiple errors updating admin channel messages")
	}
	return nil
}
