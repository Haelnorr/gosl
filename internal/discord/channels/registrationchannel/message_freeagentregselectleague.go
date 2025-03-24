package registrationchannel

import (
	"context"
	"fmt"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func registerFreeAgentSelectLeagueComponents(
	ctx context.Context,
	tx db.SafeTX,
	season *models.Season,
) (*bot.MessageContents, error) {
	leagues, err := models.GetLeagues(ctx, tx, season.ID, false)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetLeagues")
	}
	opts := []discordgo.SelectMenuOption{}
	for _, league := range *leagues {
		opts = append(opts, discordgo.SelectMenuOption{
			Label: league.Division,
			Value: fmt.Sprintf("%s", league.Division),
		})
	}
	embed := &discordgo.MessageEmbed{
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Register as Free Agent",
				Value: fmt.Sprintf(`
**Register as a Free Agent to play in %s**
Select your preferred league from the select box to apply.
**WARNING**: Clicking off the select box will send the application.
`, season.Name),
				Inline: false,
			},
		},
	}
	msgcomps := components.StringSelect(
		fmt.Sprintf("freeagent_registration_select_league"),
		"Select Preferred League",
		opts,
		1,
		1,
		false,
	)
	contents := &bot.MessageContents{
		Embed:      embed,
		Components: msgcomps,
	}
	return contents, nil
}
