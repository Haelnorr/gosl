package teamrosters

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the team rosters channel
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for team rosters channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelTeamRosters].ID {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.RBegin(timeout, "Team rosters interactions handler")
		msg := "Failed to handle interaction in team rosters channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling team rosters channel interaction")
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
			switch customID {
			case "refresh_team_rosters":
				err = handleRefresh(ctx, b, i, &ack)
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
