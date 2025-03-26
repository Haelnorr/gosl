package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/channels/teamapplications"
	"gosl/internal/discord/util"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleRegisterTeamButton(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	_, team, err := util.CheckPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}
	season, err := checkRegistrationEligibility(ctx, tx, b, team)
	if err != nil {
		if strings.Contains(err.Error(), "RF:") {
			return b.Error("Registration Failed", strings.TrimPrefix(err.Error(), "RF:"), i, *ack)
		}
		return errors.Wrap(err, "checkRegistrationEligibility")
	}
	contents, err := registerTeamComponents(ctx, tx, team, season, i.Message.ID)
	if err != nil {
		return errors.Wrap(err, "registerTeamComponents")
	}
	err = b.FollowUpComplex(contents, i, 30*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}

func handleRegisterTeamSelectLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	team, err := checkPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "checkPlayerIsManager")
	}
	season, err := checkRegistrationEligibility(ctx, tx, b, team)
	if err != nil {
		if strings.Contains(err.Error(), "RF:") {
			return b.Error("Registration Failed", strings.TrimPrefix(err.Error(), "RF:"), i, *ack)
		}
		return errors.Wrap(err, "checkRegistrationEligibility")
	}
	preferredLeague := i.MessageComponentData().Values[0]
	tr, err := team.Register(ctx, tx, season.ID, preferredLeague)
	if err != nil {
		return errors.Wrap(err, "team.Register")
	}

	regMsg, err := teamapplications.NewTeamApplicationMsg(ctx, b)
	if err != nil {
		return errors.Wrap(err, "NewTeamApplicationMsg")
	}
	contents, err := teamapplications.TeamApplicationContents(ctx, tx, tr)
	if err != nil {
		return errors.Wrap(err, "TeamApplicationContents")
	}
	err = regMsg.Send(contents)
	if err != nil {
		return errors.Wrap(err, "regMsg.Send")
	}

	updateTeamManagerPanel(ctx, tx, b, team, panelMsgID, i.User.ID)
	err = b.FollowUp(fmt.Sprintf("%s has been registered for %s!", team.Name, season.Name), i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	return nil
}

func checkRegistrationEligibility(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
	team *models.Team,
) (*models.Season, error) {
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	if season == nil {
		return nil, errors.New("RF:No Season currently active")
	}
	if !season.RegistrationOpen {
		return nil, errors.New("RF:Registration is currently closed")
	}
	now := time.Now()
	currentPlayers, err := team.Players(ctx, tx, &now, &now)
	if err != nil {
		return nil, errors.Wrap(err, "team.Players")
	}
	numPlayers := len(*currentPlayers)
	if numPlayers < 3 {
		return nil, errors.New("RF:Not enough players")
	}
	if numPlayers > 5 {
		return nil, errors.New("RF:Too many players")
	}
	if team.Color == 0x181825 {
		return nil, errors.New("RF:Team Color not set")
	}
	if team.Logo == "" {
		return nil, errors.New("RF:Team Logo not uploaded")
	}
	chanRegApp := b.Channels[models.ChannelTeamApplications]
	if chanRegApp.ID == "" {
		return nil, errors.New("Registration approvals channel not configured")
	}
	return season, nil
}
