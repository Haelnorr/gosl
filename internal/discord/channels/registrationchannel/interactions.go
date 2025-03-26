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
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.Message.ChannelID != b.Channels[models.ChannelRegistration].ID {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout, "Registration channel interaction handler")
		msg := "Failed to handle interaction in regsistration channel"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling regsistration channel interaction")

		switch i.Type {
		case discordgo.InteractionMessageComponent:
			// Handle message component interactions
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling interaction")
			switch {
			case customID == "player_registration_button":
				err = handlePlayerRegistrationButtonInteraction(ctx, tx, b, i)
			case strings.Contains(customID, "confirm_slapid_"):
				slapid := strings.TrimPrefix(customID, "confirm_slapid_")
				err = handleSteamIDConfirm(ctx, tx, b, i, slapid)
			case customID == "new_team_registration_button":
				err = handleNewTeamRegistrationButtonInteraction(ctx, tx, b, i)
			case customID == "existing_team_registration_button":
				err = handleExistingTeamRegistrationButtonInteraction(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "reregister_team_"):
				teamid := strings.TrimPrefix(customID, "reregister_team_")
				err = handleReregisterExistingTeamInteraction(ctx, tx, b, i, &ack, teamid)
			case strings.Contains(customID, "disband_team_"):
				teamid := strings.TrimPrefix(customID, "disband_team_")
				err = handleDisbandTeamInteraction(ctx, tx, b, i, &ack, teamid)
			case customID == "reregister_select_team":
				err = handleReregisterTeamSelect(ctx, tx, b, i, &ack)
			case customID == "freeagent_registration_button":
				err = handleFreeAgentRegisterButton(ctx, tx, b, i, &ack)
			case customID == "freeagent_registration_select_league":
				err = handleFreeAgentRegisterSelectLeague(ctx, tx, b, i, &ack)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		case discordgo.InteractionModalSubmit:
			// Handle modal interactions
			customID := i.ModalSubmitData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling interaction")
			switch {
			case customID == "player_reg_steam_id":
				err = handleSteamIDModalSubmit(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "player_reg_display_name_"):
				slapid := strings.TrimPrefix(customID, "player_reg_display_name_")
				err = handleDisplayNameSubmit(ctx, tx, b, i, &ack, slapid)
			case customID == "new_team_registration_details":
				err = handleNewTeamDetailsSubmit(ctx, tx, b, i, &ack)
			default:
				err = errors.New("No handler for interaction")
			}
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
