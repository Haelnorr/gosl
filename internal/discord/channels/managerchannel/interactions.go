package managerchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/permissions"
	"gosl/internal/discord/util"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleInteractions(ctx context.Context, b *util.Bot) util.Handler {
	b.Logger.Debug().Msg("Adding handler for manager channel interactions")
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		// setup the database transaction
		timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		tx, err := b.Conn.Begin(timeout)
		msg := "Failed to handle interaction in manager channel"
		if err != nil {
			b.TripleError(msg, err, s, i)
			return
		}
		defer tx.Rollback()

		// Make sure interaction happened from manager channel
		channelID, err := channels.GetChannel(ctx, tx, channels.PurposeManager)
		if err != nil {
			b.TripleError(msg, err, s, i)
			return
		}
		if i.Message.ChannelID != channelID {
			return
		}
		b.Logger.Debug().Msg("Handling manager channel interaction")

		// Check the user has permissions to do league manager things
		isLeagueManager, err := permissions.HasPermission(
			ctx, tx, s, b.Config.DiscordGuildID, i.Member, permissions.LeagueManager)
		if !isLeagueManager {
			b.Forbidden(s, i)
			return
		}

		// Check what type of interaction we are handling
		if i.Type == discordgo.InteractionMessageComponent {
			// Handle the direct interactions from message components
			switch i.MessageComponentData().CustomID {
			case "season_select":
				b.Logger.Debug().Msg("Handling season select interaction")
				err = handleSelectSeasonInteraction(ctx, tx, b, s, i)
			case "create_season_button":
				b.Logger.Debug().Msg("Handling season create button interaction")
				err = handleCreateSeasonButtonInteraction(s, i)
			case "create_season_modal":
				b.Logger.Debug().Msg("Handling season create modal interaction")
				err = handleCreateSeasonModalInteraction(ctx, tx, b, s, i)
			case "set_dates_button":
				b.Logger.Debug().Msg("Handling season create modal interaction")
				err = handleSetSeasonDatesButtonInteraction(ctx, tx, s, i)
			case "toggle_registration":
				b.Logger.Debug().Msg("Handling toggle registration interaction")
				err = handleToggleRegistrationInteraction(ctx, tx, b, s, i)
			case "select_season_leagues":
				b.Logger.Debug().Msg("Handling select leagues interaction")
				err = handleSelectLeaguesInteraction(ctx, tx, b, s, i)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		} else if i.Type == discordgo.InteractionModalSubmit {
			// Handle modal interactions
			switch i.ModalSubmitData().CustomID {
			case "create_season_modal":
				b.Logger.Debug().Msg("Handling create season modal interaction")
				err = handleCreateSeasonModalInteraction(ctx, tx, b, s, i)
			case "set_season_dates_modal":
				b.Logger.Debug().Msg("Handling set season dates modal interaction")
				err = handleSetSeasonDatesModalInteraction(ctx, tx, b, s, i)
			default:
				err = errors.New("No handler for interaction")
			}
			// error handling at end of function
		}
		// start error handling for the interaction handlers
		if err != nil {
			msg := "Failed to handle interaction in manager channel"
			b.TripleError(msg, err, s, i)
			return
		}
		tx.Commit()
	}
}
