package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/discord/directmessages"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleNewTeamRegistrationButtonInteraction(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
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
	if currentTeam != nil {
		msg := "You are already on a team. " +
			"Leave your current team or choose 'Register Existing Team'"
		return b.Error("Already on a team", msg, i, false)
	}
	steamcmp := []discordgo.MessageComponent{
		components.TextInput("team_name", "Team Name", true, "", 1, 64),
		components.TextInput("team_abbr", "Team Acronym", true, "", 3, 5),
	}

	err = b.ReplyModal("New Team Registration", "new_team_registration_details", steamcmp, i)
	if err != nil {
		return errors.Wrap(err, "b.ReplyModal")
	}
	return nil
}

func handleNewTeamDetailsSubmit(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	teamName := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	teamAbbr := i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value

	nameTaken, err := models.CheckTeamNameExists(ctx, tx, teamName)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByName")
	}
	abbrTaken, err := models.CheckTeamAbbrExists(ctx, tx, teamAbbr)
	if err != nil {
		return errors.Wrap(err, "models.GetTeamByAbbr")
	}
	msg := ""
	if nameTaken {
		msg = fmt.Sprintf("Team name '%s' is taken\n", teamName)
	}
	if abbrTaken {
		msg = msg + fmt.Sprintf("Team abbreviation '%s' is taken", teamAbbr)
	}
	if msg != "" {
		return b.Error("Cannot create team", msg, i, true)
	}

	player, err := models.GetPlayerByDiscordID(ctx, tx, i.Member.User.ID)
	if err != nil {
		return errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	team, err := models.CreateTeam(ctx, tx, teamName, teamAbbr, player.ID)
	if err != nil {
		return errors.Wrap(err, "models.CreateTeam")
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
