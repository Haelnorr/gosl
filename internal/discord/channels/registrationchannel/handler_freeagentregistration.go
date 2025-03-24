package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/freeagentapplications"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleFreeAgentRegisterButton(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	currentSeason, _, usererrmsg, err := checkFreeAgentCanRegister(ctx, tx, i.Member)
	if err != nil {
		return errors.Wrap(err, "checkFreeAgentCanRegister")
	}
	if usererrmsg != "" {
		return b.Error("Failed to register", usererrmsg, i, *ack)
	}
	contents, err := registerFreeAgentSelectLeagueComponents(ctx, tx, currentSeason)
	if err != nil {
		return errors.Wrap(err, "registerFreeAgentSelectLeagueComponents")
	}
	err = b.FollowUpComplex(contents, i, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}

func handleFreeAgentRegisterSelectLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	currentSeason, player, usererrmsg, err := checkFreeAgentCanRegister(ctx, tx, i.Member)
	if err != nil {
		return errors.Wrap(err, "checkFreeAgentCanRegister")
	}
	if usererrmsg != "" {
		return b.Error("Failed to register", usererrmsg, i, *ack)
	}
	preferredLeague := i.MessageComponentData().Values[0]
	app, err := currentSeason.RegisterFreeAgent(ctx, tx, player.ID, preferredLeague)
	if err != nil {
		return errors.Wrap(err, "currentSeason.RegisterFreeAgent")
	}
	regMsg, err := freeagentapplications.NewFreeAgentApplicationMsg(ctx, b)
	if err != nil {
		return errors.Wrap(err, "NewFreeAgentApplicationMsg")
	}
	contents, err := freeagentapplications.FreeAgentApplicationContents(ctx, tx, app)
	if err != nil {
		return errors.Wrap(err, "FreeAgentApplicationContents")
	}
	err = regMsg.Send(contents)
	if err != nil {
		return errors.Wrap(err, "regMsg.Send")
	}
	err = b.FollowUp(
		fmt.Sprintf("You have successfully applied to be a Free Agent for %s",
			currentSeason.Name), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

// checks if the player can register as a free agent. returns an empty string if
// they are allowed to register, and an error message if they are not.
func checkFreeAgentCanRegister(
	ctx context.Context,
	tx db.SafeTX,
	member *discordgo.Member,
) (*models.Season, *models.Player, string, error) {
	player, err := models.GetPlayerByDiscordID(ctx, tx, member.User.ID)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "models.GetPlayerByDiscordID")
	}
	if player == nil {
		return nil, nil, "You must register as a player first", nil
	}
	currentTeam, err := player.CurrentTeam(ctx, tx)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "player.CurrentTeam")
	}
	if currentTeam != nil {
		return nil, nil, "You are already on a team", nil
	}
	currentSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "models.GetActiveSeason")
	}
	if currentSeason == nil {
		return nil, nil, "There is no active season right now", nil
	}
	isRegistered, err := models.CheckPlayerFreeAgentRegistration(ctx, tx, player.ID, currentSeason.ID)
	if err != nil {
		return nil, nil, "", errors.Wrap(err, "models.CheckPlayerFreeAgentRegistration")
	}
	if isRegistered {
		return nil, nil, "You are already registered as a Free Agent this season", nil
	}

	// we dont check if registration is open because free agents can always register
	// if there is an active season
	return currentSeason, player, "", nil
}
