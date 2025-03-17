package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/models"

	"github.com/bwmarrin/discordgo"
)

var playerRegistration = &bot.Message{
	Label:       "Player Registration",
	Purpose:     models.MsgPlayerRegistration,
	GetContents: playerRegistrationContents,
}

func playerRegistrationContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Player Registration",
			Description: fmt.Sprintf(`
**Register as a player in the Oceanic Slapshot League!**

To register, you will need to provide the Steam ID of the account you play slapshot on. 

[Click here](%s/registration-help) for instructions on how to find your Steam ID and why it is required.
`, b.Config.TrustedHost),
			Color: 0x00ff00, // Green color
		},
		Components: []discordgo.MessageComponent{
			&discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					&discordgo.Button{
						Label:    "Player Registration",
						CustomID: "player_registration_button",
					},
				},
			},
		},
	}
	return contents, nil
}
