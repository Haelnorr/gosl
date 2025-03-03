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

func (b *Bot) setupAdminChannel(ctx context.Context) error {
	tx, err := b.conn.Begin(ctx)
	channelID, err := ensureAdminChannel(ctx, tx, b.session, b.guildID)
	if err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	b.logger.Info().Str("channel_id", channelID).Msg("Admin channel is ready")

	err = b.createAdminMessage(channelID)
	if err != nil {
		return errors.Wrap(err, "b.createAdminMessage")
	}

	b.session.AddHandler(b.handleAdminChannelInteractions(ctx))
	return nil
}

func ensureAdminChannel(
	ctx context.Context,
	tx *db.SafeTX,
	s *discordgo.Session,
	guildID string,
) (string, error) {
	channelIDs, err := getChannelsForPurpose(ctx, tx, channelAdmin)
	if err != nil {
		return "", errors.Wrap(err, "getChannelsForPurpose")
	}
	if len(channelIDs) > 0 {
		return channelIDs[0], nil
	}

	channel, err := s.GuildChannelCreate(guildID, channelAdminName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", err
	}

	err = addPurpose(ctx, tx, channel.ID, channelAdmin)
	if err != nil {
		return "", err
	}
	return channel.ID, nil
}

func (b *Bot) createAdminMessage(channelID string) error {
	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					MenuType:    discordgo.ChannelSelectMenu,
					CustomID:    "log_channel_select",
					Placeholder: "Select the channel for log output",
				},
			},
		},
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
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.SelectMenu{
					MenuType:    discordgo.RoleSelectMenu,
					CustomID:    "manager_role_select",
					Placeholder: "Select roles for manager tasks",
					MaxValues:   10,
				},
			},
		},
	}
	embed := &discordgo.MessageEmbed{
		Title: "Bot Configuration",
		Description: `
Use the dropdowns below to configure the bot settings:

**Log Channel**:
Select which channel should receive logs.

**Admin Roles**:
Select roles that can execute admin commands.

**Manager Roles**:
Select roles that can manage tasks.
`,
		Color: 0x00ff00, // Green color
	}
	messageID := getAdminChannelMessageID()
	msg := ""
	if messageID != "" {
		_, err := b.session.ChannelMessageEditComplex(&discordgo.MessageEdit{
			ID:         messageID,
			Channel:    channelID,
			Content:    &msg,
			Embeds:     &[]*discordgo.MessageEmbed{embed},
			Components: &components,
		})
		if err != nil {
			return errors.Wrap(err, "session.ChannelMessageEditComplex")
		}
	} else {
		_, err := b.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Content:    msg,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		})
		if err != nil {
			return errors.Wrap(err, "session.ChannelMessageSendComplex")
		}
	}
	return nil
}

func getAdminChannelMessageID() string {
	return "1346049817540952116"
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
