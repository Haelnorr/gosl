package registrationchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle the interactions for the registration channel components
func handleInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for registration channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Message.ChannelID != b.Channels[models.ChannelRegistration].ID {
			return
		}
		timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout)
		msg := "Failed to handle interaction in regsistration channel"
		if err != nil {
			b.TripleError(msg, err, i)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling regsistration channel interaction")

		if i.Type == discordgo.InteractionMessageComponent {
			// Handle select menu interactions
			customID := i.MessageComponentData().CustomID
			switch {
			case customID == "player_registration_button":
				b.Logger.Debug().Msg("Handling player registration button interaction")
				err = handlePlayerRegistrationButtonInteraction(ctx, tx, b, i)
			case strings.Contains(customID, "confirm_slapid_"):
				b.Logger.Debug().Msg("Handling confirm slapid interaction")
				slapid := strings.TrimPrefix(customID, "confirm_slapid_")
				err = handleSteamIDConfirm(ctx, tx, b, i, slapid)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		} else if i.Type == discordgo.InteractionModalSubmit {
			// Handle modal interactions
			customID := i.ModalSubmitData().CustomID
			switch {
			case customID == "player_reg_steam_id":
				b.Logger.Debug().Msg("Handling submit steam ID interaction")
				err = handleSteamIDModalSubmit(ctx, tx, b, i)
			case strings.Contains(customID, "player_reg_display_name_"):
				b.Logger.Debug().Msg("Handling submit display name interaction")
				slapid := strings.TrimPrefix(customID, "player_reg_display_name_")
				err = handleDisplayNameSubmit(ctx, tx, b, i, slapid)
			default:
				err = errors.New("No handler for interaction")
			}
		}
		// start error handling for interaction handlers
		if err != nil {
			msg := "Failed to handle interaction in registration channel"
			b.TripleError(msg, err, i)
			return
		}
		tx.Commit()
	}
}
