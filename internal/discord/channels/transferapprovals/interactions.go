package transferapprovals

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the transfer approvals channel
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for transfer approvals channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		transferChannel, err := b.GetChannel(models.ChannelTransferApprovals)
		if err != nil {
			b.TripleError("Interaction failed", errors.Wrap(err, "b.GetChannel"), i, false)
			return
		}
		if i.Message == nil {
			b.TripleError("Interaction failed", errors.New("InteractionCreate.Message is nil"), i, false)
			return
		}
		if i.Message.ChannelID != transferChannel.ID {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Transfer approvals interaction handler")
		msg := "Failed to handle interaction in transfer approvals channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling transfer approvals channel interaction")
		isManager, err := models.MemberHasPermission(
			ctx, tx, s, b.Config.DiscordGuildID, i.Member, models.PermLeagueManager)
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		if !isManager {
			b.Forbidden(i, ack)
			return
		}

		if i.Type == discordgo.InteractionMessageComponent {
			// Handle message component interactions
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling Interaction")
			switch {
			case strings.Contains(customID, "approve_transfer_"):
				ptiID := strings.TrimPrefix(customID, "approve_transfer_")
				err = handleApproveTransfer(ctx, tx, b, i, &ack, ptiID)
			case strings.Contains(customID, "reject_transfer_"):
				ptiID := strings.TrimPrefix(customID, "reject_transfer_")
				err = handleRejectTransfer(ctx, tx, b, i, &ack, ptiID)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		}
		// start error handling for interaction handlers
		if err != nil {
			msg := "Failed to handle interaction"
			b.TripleError(msg, err, i, ack)
			return
		}
		tx.Commit()
	}
}
