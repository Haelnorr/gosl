package managerchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleCreateSeasonButtonInteraction(
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	modalComps := []discordgo.MessageComponent{
		components.TextInput("season_id", "Season ID", true, "", 1, 5),
		components.TextInput("season_name", "Season Name", true, "", 1, 32),
	}
	err := b.ReplyModal("Create Season", "create_season_modal", modalComps, i)
	if err != nil {
		return errors.Wrap(err, "bot.ReplyModal")
	}
	return nil
}

func handleCreateSeasonModalInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	selectSeason, err := b.GetMessage(models.ChannelManager, models.MsgSelectSeason)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	selectSeason.StartUpdate(true)
	seasonID := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	seasonName := i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value

	season, err := models.CreateSeason(ctx, tx, seasonID, seasonName)
	if err != nil {
		if strings.Contains(err.Error(), "must be unique") {
			return b.Error("Error creating season", err.Error(), i, true)
		}
		return errors.Wrap(err, "models.CreateSeason")
	}
	msg := "New Season created: " + season.Name
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}

	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating season select")
		errch := make(chan error)
		go selectSeason.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update message after interaction"
				b.Logger.Warn().Err(err).Msg(msg)
				b.Log().Error(msg, err)
			}
		}
	}()
	return nil
}
