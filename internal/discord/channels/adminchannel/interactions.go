package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the admin channel components
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for admin channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelAdmin].ID {
			return
		}
		ack := false
		if i.Type == discordgo.InteractionMessageComponent {
			timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			tx, err := b.Conn.Begin(timeout)
			msg := "Failed to handle interaction in admin channel"
			if err != nil {
				b.TripleError(msg, err, i, ack)
				return
			}
			defer tx.Rollback()
			b.Logger.Debug().Msg("Handling admin channel interaction")
			isAdmin, err := models.MemberHasPermission(
				ctx, tx, s, b.Config.DiscordGuildID, i.Member, models.PermAdmin)
			if !isAdmin {
				b.Forbidden(i, ack)
				return
			}

			// Handle select menu interactions
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling interaction")
			switch customID {
			case "log_channel_select":
				err = handleSelectLogChannelInteraction(ctx, tx, b, i, &ack)
			case "admin_role_select":
				err = handleSelectAdminRolesInteraction(ctx, tx, b, i, &ack)
			case "manager_role_select":
				err = handleSelectManagerRolesInteraction(ctx, tx, b, i, &ack)
			case "registration_channel_select":
				err = handleSelectRegistrationChannelInteraction(ctx, tx, b, i, &ack)
			case "application_channel_select":
				err = handleSelectRegistrationApprovalChannelInteraction(ctx, tx, b, i, &ack)
			default:
				err = errors.New("No handler for interaction")
			}
			if err != nil {
				msg := "Failed to handle interaction"
				b.TripleError(msg, err, i, ack)
				return
			}
			tx.Commit()
		}
	}
}
