package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for manager channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelManager].ID {
			return
		}
		ack := false
		// setup the database transaction
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Manager channel interaction handler")
		msg := "Failed to handle interaction in manager channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling manager channel interaction")

		// Check the user has permissions to do league manager things
		isLeagueManager, err := models.MemberHasPermission(
			ctx, tx, s, b.Config.DiscordGuildID, i.Member, models.PermLeagueManager)
		if !isLeagueManager {
			b.Forbidden(i, ack)
			return
		}

		// Check what type of interaction we are handling
		switch i.Type {
		case discordgo.InteractionMessageComponent:
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling interaction")
			// Handle the direct interactions from message components
			switch customID {
			case "season_select":
				err = handleSelectSeasonInteraction(ctx, tx, b, i, &ack)
			case "create_season_button":
				err = handleCreateSeasonButtonInteraction(b, i)
			case "create_season_modal":
				err = handleCreateSeasonModalInteraction(ctx, tx, b, i, &ack)
			case "set_dates_button":
				err = handleSetSeasonDatesButtonInteraction(ctx, tx, b, i)
			case "toggle_registration":
				err = handleToggleRegistrationInteraction(ctx, tx, b, i, &ack)
			case "select_season_leagues":
				err = handleSelectLeaguesInteraction(ctx, tx, b, i, &ack)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		case discordgo.InteractionModalSubmit:
			customID := i.ModalSubmitData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling interaction")
			// Handle modal interactions
			switch customID {
			case "create_season_modal":
				err = handleCreateSeasonModalInteraction(ctx, tx, b, i, &ack)
			case "set_season_dates_modal":
				err = handleSetSeasonDatesModalInteraction(ctx, tx, b, i, &ack)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		}
		// start error handling for the interaction handlers
		if err != nil {
			msg := "Failed to handle interaction in manager channel"
			b.TripleError(msg, err, i, ack)
			return
		}
		tx.Commit()
	}
}
