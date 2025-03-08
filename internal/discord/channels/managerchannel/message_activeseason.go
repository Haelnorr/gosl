package managerchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Get the message contents for the show active season component
func activeSeasonComponents(
	ctx context.Context,
	b *util.Bot,
) (util.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up active season components")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := b.Conn.RBegin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()

	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	registrationButton := &discordgo.Button{
		Label:    "Open Registration",
		CustomID: "toggle_registration",
		Style:    discordgo.SuccessButton,
	}
	if season.RegistrationOpen {
		registrationButton.Label = "Close Registration"
		registrationButton.Style = discordgo.DangerButton
	}

	components := []discordgo.MessageComponent{
		&discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				registrationButton,
				&discordgo.Button{
					Label:    "Set dates",
					CustomID: "set_dates_button",
				},
			},
		},
	}
	// options := []discordgo.SelectMenuOption{
	// 	{
	// 		Label: "Open",
	// 		Value: "open",
	// 	},
	// }
	// leagueSelect := messages.StringSelect(
	// 	"select_season_leagues",
	// 	"Select Leagues",
	// 	options,
	// 	0,
	// 	3,
	// )
	// components = append(components, leagueSelect...)

	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retrieving active season components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Active Season",
				Description: fmt.Sprintf(`
**%s (%s)**

**Registration:** %s

Leagues:

Start Date: %s
Regular Season End: %s
Finals End: %s

Transfer windows:
`,
					season.Name, season.ID, season.RegistrationStatusString(),
					util.DiscordDateUntil(season.Start),
					util.DiscordDateUntil(season.RegSeasonEnd),
					util.DiscordDateUntil(season.FinalsEnd),
				),
				Color: 0x00ff00, // Green color
			},
			components
	}, nil
}

func handleSetSeasonDatesButtonInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	s *discordgo.Session,
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
	err = messages.ReplyModal("Set Season Dates", "set_season_dates_modal", components, s, i)
	if err != nil {
		return errors.Wrap(err, "messages.ReplyModal")
	}
	return nil
}

func handleSetSeasonDatesModalInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
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
	messages.ReplyEphemeral(msg, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		err := messages.UpdateChannelMessage(
			ctx,
			b,
			activeSeasonComponents,
			messages.ManagerActiveSeason,
			thisChannel,
		)
		if err != nil {
			msg := "Failed to update active season message after interaction"
			b.Logger.Warn().Err(err).
				Msg(msg)
			b.Log().Error(msg, err)
		}
	}()
	return nil
}

func handleToggleRegistrationInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	thisChannel := i.ChannelID
	b.Logger.Debug().Msg("Getting active season")
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	b.Logger.Debug().Msg("Toggling active season registration")
	err = season.ToggleRegistration(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "season.ToggleRegistration")
	}

	msg := "Registration status for %s set to %s"
	msg = fmt.Sprintf(msg, season.Name, season.RegistrationStatusString())
	b.Log().UserEvent(i.Member, msg)
	messages.ReplyEphemeral(msg, s, i, b.Logger)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		err := messages.UpdateChannelMessage(
			ctx,
			b,
			activeSeasonComponents,
			messages.ManagerActiveSeason,
			thisChannel,
		)
		if err != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update active season message after interaction")
		}
		b.Logger.Debug().Msg("Updating active season info")
	}()
	return nil
}
