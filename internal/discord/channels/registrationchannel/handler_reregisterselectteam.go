package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/discord/directmessages"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleReregisterTeamSelect(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	teamidstr := i.MessageComponentData().Values[0]
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
	currentTeam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "player.CurrentTeam")
	}
	if currentTeam != nil {
		msg = "You are already on a team!"
	}
	if msg != "" {
		return b.Error("Error re-registering team", msg, i, true)
	}
	err = player.JoinTeam(ctx, tx, team.ID)
	if err != nil {
		return errors.Wrap(err, "player.JoinTeam")
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

func teamSelectComponents(
	ctx context.Context,
	tx db.SafeTX,
	player *models.Player,
) (*bot.MessageContents, error) {
	embed := &discordgo.MessageEmbed{
		Color: 0xeb7d34,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Re-register a team",
				Value:  "Select a team from the list below to re-register",
				Inline: false,
			},
		},
	}
	teams, err := player.GetManagedTeams(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "player.GetManagedTeams")
	}
	if len(*teams) == 0 {
		return nil, errors.New("No managed teams")
	}
	opts := []discordgo.SelectMenuOption{}
	for _, team := range *teams {
		opts = append(opts, discordgo.SelectMenuOption{
			Label: team.Name,
			Value: fmt.Sprintf("%v", team.ID),
		})
	}

	msgcomps := components.StringSelect(
		"reregister_select_team",
		"Select Team",
		opts, 1, 1, false)
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}
