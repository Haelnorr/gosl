package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var teamRegistration = &bot.Message{
	Label:       "Team Registration",
	Purpose:     models.MsgTeamRegistration,
	GetContents: teamRegistrationContents,
}

func teamRegistrationContents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	activeSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	disabled := (activeSeason == nil || !activeSeason.RegistrationOpen)
	regmsg := "%s - Registration %s!"
	if activeSeason == nil {
		regmsg = fmt.Sprintf(regmsg, "No active season", "closed")
	} else if activeSeason.RegistrationOpen {
		regmsg = fmt.Sprintf(regmsg, activeSeason.Name, "open")
	} else {
		regmsg = fmt.Sprintf(regmsg, activeSeason.Name, "closed")
	}
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Team Registration",
			Description: fmt.Sprintf(`
**%s**
Register a team to play in the Oceanic Slapshot League!
`, regmsg),
			Color: 0x00ff00, // Green color
		},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Register New Team",
						CustomID: "new_team_registration_button",
						Disabled: disabled,
					},
					&discordgo.Button{
						Label:    "Register Existing Team",
						CustomID: "existing_team_registration_button",
						Disabled: disabled,
					},
				},
			},
		},
	}
	return contents, nil
}
