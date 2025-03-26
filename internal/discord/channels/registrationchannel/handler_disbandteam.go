package registrationchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleDisbandTeamInteraction(
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
		return b.Error("Error disbanding team", msg, i, true)
	}
	err = team.Disband(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "team.Disband")
	}
	contents, err := teamSelectComponents(ctx, tx, player)
	if err != nil {
		return errors.Wrap(err, "teamSelectComponents")
	}
	return b.FollowUpComplex(contents, i, 30*time.Second)
}
