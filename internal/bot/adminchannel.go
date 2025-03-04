package bot

import (
	"context"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	channelAdminName string = "gosl-bot-admin"
)

func selectLogChannelComponents(ctx context.Context, conn *db.SafeConn) (MessageContents, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	// TODO: get the values from the DB and set them in defaults
	tx.Commit()

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		return "",
			&discordgo.MessageEmbed{
				Title:       "Bot log output channel",
				Description: `Select the channel to output bot logs to`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:    discordgo.ChannelSelectMenu,
							CustomID:    "log_channel_select",
							Placeholder: "Select the channel for log output",
						},
					},
				},
			}
	}, nil
}
func selectAdminRolesComponents(ctx context.Context, conn *db.SafeConn) (MessageContents, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	// TODO: get the values from the DB and set them in defaults
	tx.Commit()

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		return "",
			&discordgo.MessageEmbed{
				Title:       "Admin roles",
				Description: `Select the roles that should have admin access`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:    discordgo.RoleSelectMenu,
							CustomID:    "admin_role_select",
							Placeholder: "Select admin roles",
							MaxValues:   10,
						},
					},
				},
			}
	}, nil
}
func selectManagerRolesComponents(ctx context.Context, conn *db.SafeConn) (MessageContents, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	// TODO: get the values from the DB and set them in defaults
	tx.Commit()

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		return "",
			&discordgo.MessageEmbed{
				Title:       "Manager roles",
				Description: `Select the roles that should have manager access`,
				Color:       0x00ff00, // Green color
			},
			[]discordgo.MessageComponent{
				&discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						&discordgo.SelectMenu{
							MenuType:    discordgo.RoleSelectMenu,
							CustomID:    "manager_role_select",
							Placeholder: "Select manager roles",
							MaxValues:   10,
						},
					},
				},
			}
	}, nil
}

func (b *Bot) handleAdminChannelInteractions(ctx context.Context) handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			channelID, err := getAdminChannel(ctx, b.conn)
			if err != nil {
				b.logger.Error().Err(err).Msg("failed to get a channel id for the admin channel")
				panic("unable to get admin channel ID")
			}
			if i.Message.ChannelID != channelID {
				return
			}
			// TODO: allow only if user is admin (discord override) or has one
			// of the set admin roles

			// Handle select menu interactions
			switch i.MessageComponentData().CustomID {
			case "log_channel_select":
				selectedChannel := i.MessageComponentData().Values[0]
				emphemeralMessage("updated log channel to"+selectedChannel, b.logger, s, i)
			case "admin_role_select":
				// TODO: update the roles in the db
				_ = i.MessageComponentData().Values
				emphemeralMessage("updated admin roles", b.logger, s, i)
			case "manager_role_select":
				// TODO: update the roles in the db
				_ = i.MessageComponentData().Values
				emphemeralMessage("updated admin roles", b.logger, s, i)
			}
		}
	}
}

func (b *Bot) setupAdminChannel(ctx context.Context) error {
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
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	defer tx.Commit()
	channelIDs, err := getChannelsForPurpose(ctx, tx, channelAdmin)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	for _, channelID := range channelIDs {
		if exists := b.checkChannelExists(channelID); exists {
			return channelIDs[0], nil
		} else {
			removeChannelPurpose(ctx, tx, channelID, channelAdmin)
		}
	}

	channel, err := b.session.GuildChannelCreate(
		b.guildID, channelAdminName, discordgo.ChannelTypeGuildText)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "s.GuildChannelCreate")
	}

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
	components, err := selectLogChannelComponents(ctx, b.conn)
	if err != nil {
		return errors.Wrap(err, "selectLogChannelComponents")
	}
	err = b.updateAdminChannelMessage(
		ctx, channelID, messageAdminSelectLogChannel, components)
	if err != nil {
		return errors.Wrap(err, "updateAdminChannelMessage: SelectLogChannel")
	}

	// Select admin roles
	components, err = selectAdminRolesComponents(ctx, b.conn)
	if err != nil {
		return errors.Wrap(err, "selectLogChannelComponents")
	}
	err = b.updateAdminChannelMessage(
		ctx, channelID, messageAdminSelectAdminRoles, components)
	if err != nil {
		return errors.Wrap(err, "updateAdminChannelMessage: SelectAdminRoles")
	}

	// Select manager roles
	components, err = selectManagerRolesComponents(ctx, b.conn)
	if err != nil {
		return errors.Wrap(err, "selectLogChannelComponents")
	}
	err = b.updateAdminChannelMessage(
		ctx, channelID, messageAdminSelectManagerRoles, components)
	if err != nil {
		return errors.Wrap(err, "updateAdminChannelMessage: SelectManagerRoles")
	}
	return nil
}

func (b *Bot) updateAdminChannelMessage(
	ctx context.Context,
	adminChannelID string,
	purpose uint16,
	contents MessageContents,
) error {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.conn.Begin(timeout)
	defer tx.Commit()
	messageID, channelID, err := getMessageForPurpose(
		ctx, tx, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "getMessageForPurpose")
	}
	if messageID != "" && channelID != "" {
		if exists := b.checkMessageExists(messageID, channelID); exists {
			err = b.editStaticMessage(contents, messageID, channelID)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "b.editStaticMessage")
			}
			return nil
		}
	}
	messageID, err = b.createStaticMessage(contents, adminChannelID)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "b.createStaticMessage")
	}
	err = addMessagePurpose(ctx, tx, messageID, adminChannelID, purpose)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "addMessagePurpose")
	}
	return nil
}

func getAdminChannel(ctx context.Context, conn *db.SafeConn) (string, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := conn.Begin(timeout)
	if err != nil {
		return "", errors.Wrap(err, "conn.Begin")
	}
	channelIDs, err := getChannelsForPurpose(ctx, tx, channelAdmin)
	if err != nil {
		tx.Rollback()
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	if len(channelIDs) == 0 {
		tx.Rollback()
		return "", errors.New("No admin channels found")
	}
	tx.Commit()
	return channelIDs[0], nil
}
