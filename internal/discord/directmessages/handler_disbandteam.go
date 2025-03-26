package directmessages

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleDisbandTeam(
	ctx context.Context,
	tx *db.SafeWTX,
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
		return errors.Wrap(err, "util.CheckPlayerIsManager")
	}
	contents := disbandTeamComponents(team, i.Message.ID)
	err = b.FollowUpComplex(contents, i, 20*time.Second)
	if err != nil {
		return errors.Wrap(err, "b.FollowUpComplex")
	}
	return nil
}

func handleDisbandTeamConfirm(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
	panelMsgID string,
) error {
	b.Acknowledge(i, ack)
	_, team, err := util.CheckPlayerIsManager(ctx, tx, i.User.ID)
	if err != nil {
		if strings.Contains(err.Error(), "VE:") {
			return b.Error("Interaction failed", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "util.CheckPlayerIsManager")
	}
	err = team.Disband(ctx, tx)
	if err != nil {
		if err.Error() == "Team cannot be disbanded as they are in an active league" {
			return b.Error("Failed to disband team", err.Error(), i, *ack)
		}
		return errors.Wrap(err, "team.Disband")
	}
	panelMsg, err := b.GetDirectMessage(
		panelMsgID,
		i.User.ID,
		"Team Manager Panel",
		5*time.Minute,
		false,
	)
	if err != nil {
		return errors.Wrap(err, "b.GetDirectMessage")
	}
	contents, err := TeamManagerComponents(ctx, tx, b, team)
	if err != nil {
		return errors.Wrap(err, "TeamManagerComponents")
	}
	err = panelMsg.Expire(contents)
	if err != nil {
		return errors.Wrap(err, "panelMsg.Expire")
	}
	err = b.FollowUp(fmt.Sprintf("%s has been disbanded", team.Name), i)
	return nil
}
