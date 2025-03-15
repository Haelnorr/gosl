package managerchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var activeSeasonInfo = &bot.Message{
	Label:       "Active Season Info",
	Purpose:     models.MsgActiveSeason,
	GetContents: activeSeasonComponents,
}

// Get the message contents for the show active season component
func activeSeasonComponents(
	ctx context.Context,
	b *bot.Bot,
) (bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up active season components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
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
	comps := []discordgo.MessageComponent{}

	leagues, err := models.GetLeagues(ctx, tx, season.ID)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetLeagues")
	}
	if season.ID != "NOACTIVESEASON" {
		comps = []discordgo.MessageComponent{
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
		options, err := getLeagueOptions(leagues)
		if err != nil {
			return nil, errors.Wrap(err, "getLeagueOptions")
		}
		leagueSelect := components.StringSelect(
			"select_season_leagues",
			"Select Leagues",
			options,
			0,
			3,
		)
		comps = append(comps, leagueSelect...)
	}
	tx.Commit()
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retrieving active season components")
		embed :=
			&discordgo.MessageEmbed{
				Title: "Active Season",
				Description: fmt.Sprintf(`
                    **%s (%s)**

                    **Registration:** %s

                    **Leagues:** %s

                    Start Date: %s
                    Regular Season End: %s
                    Finals End: %s

                    Transfer windows:
                    `,
					season.Name, season.ID, season.RegistrationStatusString(),
					func() string {
						msg := ""
						for i, league := range *leagues {
							msg = msg + league.Division
							if i+1 < len(*leagues) {
								msg = msg + ", "
							}
						}
						return msg
					}(),
					bot.DiscordDateUntil(season.Start),
					bot.DiscordDateUntil(season.RegSeasonEnd),
					bot.DiscordDateUntil(season.FinalsEnd),
				),
				Color: 0x00ff00, // Green color
			}
		return "", embed, comps
	}, nil
}

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
) error {
	activeSeasonInfo := b.Channels[models.ChannelManager].Messages[models.MsgActiveSeason]
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
	b.Reply(msg, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		errch := make(chan error)
		go activeSeasonInfo.Update(ctx, errch)
		if <-errch != nil {
			msg := "Failed to update active season message after interaction"
			b.Logger.Warn().Err(err).
				Msg(msg)
			b.Log().Error(msg, err)
		}
		close(errch)
	}()
	return nil
}

func handleToggleRegistrationInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	activeSeasonInfo := b.Channels[models.ChannelManager].Messages[models.MsgActiveSeason]
	activeSeasonInfo.StartUpdate(false)
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
	b.Reply(msg, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		errch := make(chan error)
		go activeSeasonInfo.Update(ctx, errch)
		if <-errch != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update active season message after interaction")
		}
		b.Logger.Debug().Msg("Updating active season info")
		close(errch)
	}()
	return nil
}

func handleSelectLeaguesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	activeSeasonInfo := b.Channels[models.ChannelManager].Messages[models.MsgActiveSeason]
	activeSeasonInfo.StartUpdate(false)
	b.Logger.Debug().Msg("Getting active season")
	season, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return errors.Wrap(err, "models.GetActiveSeason")
	}
	leagues := i.MessageComponentData().Values
	err = models.SetLeagues(ctx, tx, season.ID, leagues)
	if err != nil {
		return errors.Wrap(err, "models.SetLeagues")
	}

	msg := "Leagues updated for %s:\n"
	for _, league := range leagues {
		msg = msg + " - " + league + "\n"
	}
	msg = fmt.Sprintf(msg, season.Name)
	b.Log().UserEvent(i.Member, msg)
	b.Reply(msg, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		b.Logger.Debug().Msg("Updating active season message")
		errch := make(chan error)
		go activeSeasonInfo.Update(ctx, errch)
		if <-errch != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update active season message after interaction")
		}
		b.Logger.Debug().Msg("Updating active season info")
		close(errch)
	}()
	return nil
}

func getLeagueOptions(
	leagues *[]models.League,
) ([]discordgo.SelectMenuOption, error) {
	allLeagues := map[string]*discordgo.SelectMenuOption{

		"Open": {
			Label: "Open",
			Value: "Open",
		},
		"IM": {
			Label: "Intermediate",
			Value: "IM",
		},
		"Pro": {
			Label: "Pro",
			Value: "Pro",
		},
	}
	options := []discordgo.SelectMenuOption{}
	for _, league := range *leagues {
		allLeagues[league.Division].Default = true
	}
	for _, option := range allLeagues {
		options = append(options, *option)
	}
	return options, nil
}
