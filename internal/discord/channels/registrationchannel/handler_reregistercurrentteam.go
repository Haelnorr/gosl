package registrationchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/directmessages"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleReregisterExistingTeamInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	teamidstr string,
) error {
	b.Acknowledge(i, ack)
	teamid, err := strconv.ParseUint(teamidstr, 10, 0)
	if err != nil {
		return errors.Wrap(err, "strconv.ParseUint")
	}
	player, err := models.GetPlayerByDiscordID(ctx, tx, i.Member.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	team, err := models.GetTeamByID(ctx, tx, uint16(teamid))
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByID")
	}
	msg := ""
	if player == nil {
		msg = "Not registered as a player"
	}
	if team == nil {
		msg = "Team not found"
	}
	if team.ManagerID != player.ID {
		msg = "You are not the manager of this team!"
	}
	if msg != "" {
		return b.Error("Error re-registering team", msg, i, true)
	}
	contents, err := directmessages.TeamManagerComponents(ctx, tx, b, team)
	if err != nil {
		return errors.Wrap(err, "components.TeamManagerComponents")
	}
	dm := bot.NewDirectMessage(
		"Team manager panel",
		i.Member.User.ID,
		5*time.Minute,
		false,
		b,
	)
	err = dm.Send(contents)
	if err != nil {
		return errors.Wrap(err, "dm.Send")
	}
	err = b.FollowUp("Team registration started, check your DM's to continue", i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}
