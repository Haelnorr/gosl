package managerchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func handleSetSeasonDatesButtonInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	startDate := models.DateStr(season.Start)
	endDate := models.DateStr(season.RegSeasonEnd)
	finalsDate := models.DateStr(season.FinalsEnd)
	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.TextInput{
					CustomID:    "season_start_date",
					Label:       "Start Date (DD/MM/YYYY)",
					Style:       discordgo.TextInputShort,
					Placeholder: "DD/MM/YYYY",
					Required:    false,
					Value:       startDate,
				},
			},
		},
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.TextInput{
					CustomID:    "season_regend_date",
					Label:       "Regulation End Date (DD/MM/YYYY)",
					Style:       discordgo.TextInputShort,
					Placeholder: "DD/MM/YYYY",
					Required:    false,
					Value:       endDate,
				},
			},
		},
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				&discordgo.TextInput{
					CustomID:    "season_finalsend_date",
					Label:       "Finals End Date (DD/MM/YYYY)",
					Style:       discordgo.TextInputShort,
					Placeholder: "DD/MM/YYYY",
					Required:    false,
					Value:       finalsDate,
				},
			},
		},
	}
	err = b.ReplyModal("Set Season Dates", "set_season_dates_modal", components, i)
	if err != nil {
		return errors.Wrap(err, "messages.ReplyModal")
	}
	return nil
}

func handleSetSeasonDatesModalInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	activeSeasonInfo, err := b.GetMessage(models.ChannelManager, models.MsgActiveSeason)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	activeSeasonInfo.StartUpdate(false)
	startDate := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	regEndDate := i.ModalSubmitData().Components[1].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	finalsEndDate := i.ModalSubmitData().Components[2].(*discordgo.ActionsRow).
		Components[0].(*discordgo.TextInput).Value
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	err = season.SetDates(ctx, tx, startDate, regEndDate, finalsEndDate, b.Config.Locale)
	if err != nil {
		return errors.Wrap(err, "season.SetDates")
	}

	msg := `
**Dates updated for %s:**
Start: %s
Regular Season End: %s
Finals End: %s`
	msg = fmt.Sprintf(msg, season.Name, season.Start, season.RegSeasonEnd, season.FinalsEnd)
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		errch := make(chan error)
		go activeSeasonInfo.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to message after interaction"
				b.Logger.Warn().Err(err).Msg(msg)
				b.Log().Error(msg, err)
			}
		}
	}()
	return nil
}
