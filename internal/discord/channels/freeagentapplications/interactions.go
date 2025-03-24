package freeagentapplications

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the free agent applications channel
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for free agent applications channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelFreeAgentApplications].ID {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Free agent applications interaction handler")
		msg := "Failed to handle interaction in free agent applications channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling free agent applications channel interaction")
		isManager, err := models.MemberHasPermission(
			ctx, tx, s, b.Config.DiscordGuildID, i.Member, models.PermLeagueManager)
		if !isManager {
			b.Forbidden(i, ack)
			return
		}

		if i.Type == discordgo.InteractionMessageComponent {
			// Handle message component interactions
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling Interaction")
			switch {
			case strings.Contains(customID, "approve_freeagent_application_"):
				applicationID := strings.TrimPrefix(customID, "approve_freeagent_application_")
				err = handleApproveFreeAgentApplication(ctx, tx, b, i, &ack, applicationID)
			case strings.Contains(customID, "reject_freeagent_application_"):
				applicationID := strings.TrimPrefix(customID, "reject_freeagent_application_")
				err = handleRejectFreeAgentApplication(ctx, tx, b, i, &ack, applicationID)
			case strings.Contains(customID, "place_freeagent_league_select_"):
				applicationID := strings.TrimPrefix(customID, "place_freeagent_league_select_")
				err = handlePlaceFreeAgentLeagueSelect(ctx, tx, b, i, &ack, applicationID)
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
