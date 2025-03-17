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

func invitePlayersComponents(
	ctx context.Context,
	tx db.SafeTX,
	team *models.Team,
	messageID string,
) (*bot.MessageContents, error) {
	players, err := models.GetInviteablePlayers(ctx, tx, team.ID)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetInviteablePlayers")
	}
	opts := []discordgo.SelectMenuOption{}
	for _, player := range *players {
		opts = append(opts, discordgo.SelectMenuOption{
			Label: player.Name,
			Value: player.DiscordID,
		})
	}
	maxPlayers := 5
	eligiblePlayers := len(opts)
	if eligiblePlayers == 0 {
		return nil, errors.New("No eligible players")
	}
	if eligiblePlayers < 5 {
		maxPlayers = eligiblePlayers
	}
	embed := &discordgo.MessageEmbed{
		Color: team.Color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name: "Invite eligble players",
				Value: `
Select players to invite from the list.
**WARNING**: Clicking off the select box will send the invitations.

*Only registered players not currently on a team can be invited.*
`,
				Inline: false,
			},
		},
	}
	comps := components.StringSelect(
		fmt.Sprintf("invite_selected_players_%s", messageID),
		"Invite players", opts, 0, maxPlayers, false)
	return &bot.MessageContents{
		Embed:      embed,
		Components: comps,
	}, nil
}
