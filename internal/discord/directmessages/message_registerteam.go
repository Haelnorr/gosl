package directmessages

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

func registerTeamComponents(
	ctx context.Context,
	tx db.SafeTX,
	team *models.Team,
	season *models.Season,
	messageID string,
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
				Name: "Register Team",
				Value: fmt.Sprintf(`
**Register %s to play in %s**
Select your preferred league from the select box to apply.
**WARNING**: Clicking off the select box will send the application.
`, team.Name, season.Name),
				Inline: false,
			},
		},
	}
	msgcomps := components.StringSelect(
		fmt.Sprintf("register_team_select_league_%s", messageID),
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
