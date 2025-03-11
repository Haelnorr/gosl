package adminchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/permissions"
	"gosl/internal/discord/util"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the admin channel components
func handleInteractions(ctx context.Context, b *util.Bot) util.Handler {
	b.Logger.Debug().Msg("Adding handler for admin channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionMessageComponent {
			timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			tx, err := b.Conn.Begin(timeout)
			msg := "Failed to handle interaction in admin channel"
			if err != nil {
				b.TripleError(msg, err, s, i)
				return
			}
			defer tx.Rollback()
			channelID, err := channels.GetChannel(ctx, tx, channels.PurposeAdmin)
			if err != nil {
				b.TripleError(msg, err, s, i)
				return
			}
			if i.Message.ChannelID != channelID {
				return
			}
			b.Logger.Debug().Msg("Handling admin channel interaction")
			isAdmin, err := permissions.HasPermission(
				ctx, tx, s, b.Config.DiscordGuildID, i.Member, permissions.Admin)
			if !isAdmin {
				b.Forbidden(s, i)
				return
			}

			// Handle select menu interactions
			switch i.MessageComponentData().CustomID {
			case "log_channel_select":
				b.Logger.Debug().Msg("Handling log channel select interaction")
				err = handleSelectLogChannelInteraction(ctx, tx, b, s, i)
			case "admin_role_select":
				b.Logger.Debug().Msg("Handling admin roles select interaction")
				err = handleSelectAdminRolesInteraction(ctx, tx, b, s, i)
			case "manager_role_select":
				b.Logger.Debug().Msg("Handling manager roles select interaction")
				err = handleSelectManagerRolesInteraction(ctx, tx, b, s, i)
			default:
				err = errors.New("No handler for interaction")
			}
			if err != nil {
				msg := "Failed to handle interaction in admin channel"
				b.TripleError(msg, err, s, i)
				return
			}
			tx.Commit()
		}
	}
}
