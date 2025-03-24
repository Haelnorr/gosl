package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"time"

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
	b *bot.Bot,
) (*bot.MessageContents, error) {
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.RBegin(timeout, "teamRegistrationContents()")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.RBegin")
	}
	defer tx.Rollback()

	activeSeason, err := models.GetActiveSeason(ctx, tx)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetActiveSeason")
	}
	disabled := false
	if activeSeason == nil {
		disabled = true
	}
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
