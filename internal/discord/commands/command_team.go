package commands

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/directmessages"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func cmdTeam(ctx context.Context, b *bot.Bot) *Command {
	return &Command{
		Name:        "team",
		Description: "View team info",
		Handler:     handleTeam(ctx, b),
	}
}

func handleTeam(
	ctx context.Context,
	b *bot.Bot,
) bot.Handler {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		b.Acknowledge(i, nil)
		timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		tx, err := b.Conn.RBegin(timeout, "Handle /team command")
		if err != nil {
			b.TripleError("Unexpected error", err, i, true)
			return
		}
		defer tx.Rollback()
		inDMs := false
		var discordID string
		if i.User == nil {
			discordID = i.Member.User.ID
		} else {
			discordID = i.User.ID
			inDMs = true
		}
		player, err := models.GetPlayerByDiscordID(ctx, tx, discordID)
		if err != nil {
			b.TripleError("Unexpected error", errors.Wrap(err, "models.GetPlayerByDiscordID"), i, true)
			return
		}
		if player == nil {
			b.Error(
				"Unregistered player",
				"You are not registered as a player. Please register to use this command",
				i, true,
			)
			return
		}
		pt, err := player.CurrentTeam(ctx, tx)
		if err != nil {
			b.TripleError("Unexpected error", errors.Wrap(err, "player.CurrentTeam"), i, true)
			return
		}
		if pt == nil {
			b.Error(
				"Not on a team",
				"You are not currently a member of a team. Join or create one to use this command",
				i, true,
			)
			return
		}
		team, err := models.GetTeamByID(ctx, tx, pt.TeamID)
		if err != nil {
			b.TripleError("Unexpected error", errors.Wrap(err, "models.GetTeamByID"), i, true)
			return
		}
		var contents *bot.MessageContents
		if team.ManagerID == player.ID {
			contents, err = directmessages.TeamManagerComponents(ctx, tx, b, team)
			if err != nil {
				b.TripleError("Unexpected error", errors.Wrap(err, "TeamManagerComponents"), i, true)
				return
			}
		} else {
			contents, err = directmessages.TeamPlayerComponents(ctx, tx, team)
			if err != nil {
				b.TripleError("Unexpected error", errors.Wrap(err, "teamPlayerComponents"), i, true)
				return
			}
		}
		dm := bot.NewDirectMessage(
			"Team view panel",
			discordID,
			5*time.Minute,
			false,
			b,
		)
		err = dm.Send(contents)
		if err != nil {
			b.TripleError("Unexpected error", errors.Wrap(err, "dm.Send"), i, true)
			return
		}
		msg := "Check your DM's"
		if inDMs {
			msg = "Viewing current team"
		}
		err = b.FollowUp(msg, i)
		if err != nil {
			b.Logger.Error().Err(err).Msg("Failed to reply to interaction")
		}
	}
}
