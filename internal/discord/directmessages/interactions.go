package directmessages

import (
	"context"
	"errors"
	"gosl/internal/discord/bot"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func HandleDMInteractions(ctx context.Context, b *bot.Bot) bot.Handler {
	b.Logger.Debug().Msg("Adding handler for direct message interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Type == discordgo.InteractionApplicationCommand {
			return
		}
		if i.User == nil {
			return
		}
		ack := false
		timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout)
		msg := "Failed to handle interaction in direct messages"
		if err != nil {
			b.TripleError(msg, err, i, ack)
			return
		}
		defer tx.Rollback()
		b.Logger.Debug().Msg("Handling direct message interaction")
		if i.Type == discordgo.InteractionMessageComponent {
			// Handle message component interactions
			customID := i.MessageComponentData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling Interaction")
			switch {
			case customID == "invite_players_button":
				err = handleInvitePlayersInteraction(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "invite_selected_players_"):
				panelMsgID := strings.TrimPrefix(customID, "invite_selected_players_")
				err = handleInviteSelectedPlayersInteraction(ctx, tx, b, i, &ack, panelMsgID)
			case customID == "remove_players_button":
				err = handleRemovePlayersButton(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "remove_player_"):
				args := strings.Split(strings.TrimPrefix(customID, "remove_player_"), "_")
				playerID := args[0]
				panelMsgID := args[1]
				err = handleRemovePlayerInteraction(ctx, tx, b, i, &ack, playerID, panelMsgID)
			case strings.Contains(customID, "revoke_invite_"):
				args := strings.Split(strings.TrimPrefix(customID, "revoke_invite_"), "_")
				inviteID := args[0]
				panelMsgID := args[1]
				err = handleRevokeInvite(ctx, tx, b, i, &ack, inviteID, panelMsgID)
			case customID == "disband_team_button":
				err = handleDisbandTeam(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "disband_team_confirm_"):
				panelMsgID := strings.TrimPrefix(customID, "disband_team_confirm_")
				err = handleDisbandTeamConfirm(ctx, tx, b, i, &ack, panelMsgID)
			case customID == "set_color_button":
				err = handleColorButton(b, i)
			case customID == "register_team_button":
				err = handleRegisterTeamButton(ctx, tx, b, i, &ack)
			case strings.Contains(customID, "register_team_select_league_"):
				panelMsgID := strings.TrimPrefix(customID, "register_team_select_league_")
				err = handleRegisterTeamSelectLeague(ctx, tx, b, i, &ack, panelMsgID)
			case strings.Contains(customID, "accept_invite_"):
				args := strings.Split(strings.TrimPrefix(customID, "accept_invite_"), "_")
				inviteID := args[0]
				panelMsgID := args[1]
				err = handleAcceptInvite(ctx, tx, b, i, &ack, inviteID, panelMsgID)
			case strings.Contains(customID, "reject_invite_"):
				args := strings.Split(strings.TrimPrefix(customID, "reject_invite_"), "_")
				inviteID := args[0]
				panelMsgID := args[1]
				err = handleRejectInvite(ctx, tx, b, i, &ack, inviteID, panelMsgID)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		} else if i.Type == discordgo.InteractionModalSubmit {
			// Handle modal interactions
			customID := i.ModalSubmitData().CustomID
			b.Logger.Debug().Str("custom_id", customID).Msg("Handling Interaction")
			switch {
			case strings.Contains(customID, "set_color_modal_"):
				panelMsgID := strings.TrimPrefix(customID, "set_color_modal_")
				err = handleSetTeamColor(ctx, tx, b, i, &ack, panelMsgID)
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
