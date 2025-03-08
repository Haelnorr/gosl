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
		if i.Type == discordgo.InteractionMessageComponent {
			timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			tx, err := b.Conn.Begin(timeout)
			if err != nil {
				msg := "Failed to start database transaction"
				b.Logger.Error().Err(err).Msg(msg)
				b.Error("Database error occured", &msg, s, i)
				return
			}
			defer tx.Rollback()
			channelID, err := channels.GetChannel(ctx, tx, channels.PurposeManager)
			if err != nil {
				b.Logger.Error().Err(err).Msg("failed to get a channel id for the admin channel")
				panic("unable to get manager channel ID")
			}
			if i.Message.ChannelID != channelID {
				return
			}
			b.Logger.Debug().Msg("Handling manager channel interaction")
			isLeagueManager, err := permissions.HasPermission(
				ctx, tx, s, b.GuildID, i.Member, permissions.LeagueManager)
			if !isLeagueManager {
				msg := "You do not have permission for this action"
				b.Error("Forbidden", &msg, s, i)
				return
			}

			// Handle select menu interactions
			switch i.MessageComponentData().CustomID {
			case "season_select":
				b.Logger.Debug().Msg("Handling season select interaction")
				err = handleSelectSeasonInteraction(ctx, tx, b, s, i)
			default:
				err = errors.New("No handler for interaction")
			}
			if err != nil {
				msg := "Interaction failed"
				smsg := err.Error()
				b.Logger.Error().Err(err).Msg(msg)
				b.Error(msg, &smsg, s, i)
				return
			}
			tx.Commit()
		}
	}
}
