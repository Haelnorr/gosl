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

var freeagentRegistration = &bot.Message{
	Label:       "Free Agent Registration",
	Purpose:     models.MsgFreeAgentRegistration,
	GetContents: freeAgentRegistrationContents,
}

func freeAgentRegistrationContents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	activeSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	disabled := activeSeason == nil
	regmsg := "%s - Registration %s!"
	if activeSeason == nil {
		regmsg = fmt.Sprintf(regmsg, "No active season", "closed")
	} else {
		regmsg = fmt.Sprintf(regmsg, activeSeason.Name, "open")
	}
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Free Agent Registration",
			Description: fmt.Sprintf(`
**%s**
Register as a Free Agent in the Oceanic Slapshot League!
`, regmsg),
			Color: 0x00ff00, // Green color
		},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Free Agent Registration",
						CustomID: "freeagent_registration_button",
						Disabled: disabled,
					},
				},
			},
		},
	}
	return contents, nil
}
