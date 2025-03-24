package teamapplications

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the team applications channel
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for team applications channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelTeamApplications].ID {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Team applications interaction handler")
		msg := "Failed to handle interaction in team applications channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling team applications channel interaction")
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
			case strings.Contains(customID, "refresh_team_application_"):
				applicationID := strings.TrimPrefix(customID, "refresh_team_application_")
				err = handleRefreshTeamApplication(ctx, tx, b, i, &ack, applicationID)
			case strings.Contains(customID, "approve_team_application_"):
				applicationID := strings.TrimPrefix(customID, "approve_team_application_")
				err = handleApproveTeamApplication(ctx, tx, b, i, &ack, applicationID)
			case strings.Contains(customID, "reject_team_application_"):
				applicationID := strings.TrimPrefix(customID, "reject_team_application_")
				err = handleRejectTeamApplication(ctx, tx, b, i, &ack, applicationID)
			case strings.Contains(customID, "place_team_league_select_"):
				applicationID := strings.TrimPrefix(customID, "place_team_league_select_")
				err = handlePlaceTeamLeagueSelect(ctx, tx, b, i, &ack, applicationID)
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
