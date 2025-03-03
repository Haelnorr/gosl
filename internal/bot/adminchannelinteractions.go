package bot

import (
	"context"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
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
			isAdmin, err := hasPermission(ctx, tx, s, b.guildID, i.Member.User, permissionAdmin)
			if !isAdmin {
				msg := "You do not have permission for this action"
				errorResponse("Forbidden", &msg, b.files, s, i)
				return
			}

			// Handle select menu interactions
			switch i.MessageComponentData().CustomID {
			case "log_channel_select":
				b.logger.Debug().Msg("Handling log channel select interaction")
				err = b.handleSelectLogChannelInteraction(ctx, tx, s, i)
			case "admin_role_select":
				b.logger.Debug().Msg("Handling admin roles select interaction")
				err = b.handleSelectAdminRolesInteraction(ctx, tx, s, i)
			case "manager_role_select":
				b.logger.Debug().Msg("Handling manager roles select interaction")
				err = b.handleSelectManagerRolesInteraction(ctx, tx, s, i)
			default:
				err = errors.New("No handler for interaction")
			}
			if err != nil {
				msg := "Interaction failed"
				smsg := err.Error()
				b.logger.Error().Err(err).Msg(msg)
				errorResponse(msg, &smsg, b.files, s, i)
				return
			}
			tx.Commit()
		}
	}
}

func (b *Bot) handleSelectLogChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	selectedChannel := i.MessageComponentData().Values[0]
	err := setChannelPurpose(ctx, tx, selectedChannel, channelLog)
	if err != nil {
		return errors.Wrap(err, "setChannelPurpose (log channel)")
	}
	b.ephemeralReply("Updated log channel to "+selectedChannel, s, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.logger.Debug().Msg("Updating log channel select")
		err = b.updateChannelMessage(
			ctx,
			b.selectLogChannelContents,
			messageAdminSelectLogChannel,
			thisChannel,
		)
		if err != nil {
			b.logger.Warn().Err(err).
				Msg("Failed to update select log channel message after interaction")
		}
	}()
	return nil
}

func (b *Bot) handleSelectAdminRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	roles := i.MessageComponentData().Values
	err := setRolesForPermission(ctx, tx, roles, permissionAdmin)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (admin)")
	}
	b.ephemeralReply("updated admin roles", s, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.logger.Debug().Msg("Updating log channel select")
		err = b.updateChannelMessage(
			ctx,
			b.selectAdminRolesContents,
			messageAdminSelectAdminRoles,
			thisChannel,
		)
		if err != nil {
			b.logger.Warn().Err(err).
				Msg("Failed to update select log channel message after interaction")
		}
	}()
	return nil
}

func (b *Bot) handleSelectManagerRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	roles := i.MessageComponentData().Values
	err := setRolesForPermission(ctx, tx, roles, permissionLeagueManager)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (manager)")
	}
	b.ephemeralReply("updated league manager roles", s, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.logger.Debug().Msg("Updating log channel select")
		err = b.updateChannelMessage(
			ctx,
			b.selectManagerRolesContents,
			messageAdminSelectManagerRoles,
			thisChannel,
		)
		if err != nil {
			b.logger.Warn().Err(err).
				Msg("Failed to update select log channel message after interaction")
		}
	}()
	return nil
}
