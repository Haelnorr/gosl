package managerchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
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
) (*bot.MessageContents, error) {
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
	if season == nil {
		season = &models.Season{
			ID:               "NOACTIVESEASON",
			Name:             "No active season",
			Active:           false,
			RegistrationOpen: false,
		}
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

	leagues, err := models.GetLeagues(ctx, tx, season.ID, true)
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
			false,
		)
		comps = append(comps, leagueSelect...)
	}
	tx.Commit()
	embed := &discordgo.MessageEmbed{
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
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: comps,
	}
	return contents, nil
}
