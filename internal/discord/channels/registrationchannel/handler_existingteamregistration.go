package registrationchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleExistingTeamRegistrationButtonInteraction(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.Member.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player == nil {
		return b.Error("Unregistered Player", "You must register as a player to register a team", i, false)
	}
	currentTeam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	var contents *bot.MessageContents
	if currentTeam != nil {
		team, err := models.GetTeamByID(ctx, tx, currentTeam.TeamID)
		if err != nil {
			return errors.Wrap(err, "models.GetTeamByID")
		}
		if team.ManagerID != player.ID {

			msg := "You are already on a team. " +
				"Leave your current team or choose 'Register Existing Team'"
			return b.Error("Already on a team", msg, i, false)
		}
		contents = reregisterTeamComponents(team)
	} else {
		contents, err = teamSelectComponents(ctx, tx, player)
		if err != nil {
			if err.Error() == "No managed teams" {
				return b.Error("Unable to re-register team", "You have not managed any teams previously. Please create a new team.", i, *ack)
			}
			return errors.Wrap(err, "teamSelectComponents")
		}
	}
	err = b.FollowUpComplex(contents, i, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}
