package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectFARoles = &bot.Message{
	Label:       "Select Free Agent Roles",
	Purpose:     models.MsgSelectFARoles,
	GetContents: selectFARolesContents,
}

// Get the message contents for the select roles message
func selectFARolesContents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select free agent roles message")
	proFARoleSelect, err := getRoleSelect(ctx, tx, models.PermProFreeAgent,
		"pro_freeagent_role_select", "Select Pro Free Agent role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	imFARoleSelect, err := getRoleSelect(ctx, tx, models.PermIMFreeAgent,
		"im_freeagent_role_select", "Select IM Free Agent role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	openFARoleSelect, err := getRoleSelect(ctx, tx, models.PermOpenFreeAgent,
		"open_freeagent_role_select", "Select Open Free Agent role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}

	msgcomps := []discordgo.MessageComponent{}
	msgcomps = append(msgcomps, proFARoleSelect...)
	msgcomps = append(msgcomps, imFARoleSelect...)
	msgcomps = append(msgcomps, openFARoleSelect...)

	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Free Agents Role select",
			Description: `
**Free Agent roles**
Select the roles that should be used for Free Agents
`,

			Color: 0x00ff00, // Green color
		},
		Components: msgcomps,
	}
	return contents, nil
}
