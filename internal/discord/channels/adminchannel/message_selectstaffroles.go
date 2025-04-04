package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectStaffRoles = &bot.Message{
	Label:       "Select Staff Roles",
	Purpose:     models.MsgSelectStaffRoles,
	GetContents: selectStaffRolesContents,
}

// Get the message contents for the select roles message
func selectStaffRolesContents(
	ctx context.Context,
	tx db.SafeTX,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select staff roles message")
	adminRoleSelect, err := getRoleSelect(ctx, tx, models.PermAdmin,
		"admin_role_select", "Select Admin roles", 0, 10)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	leagueMgrRoleSelect, err := getRoleSelect(ctx, tx, models.PermLeagueManager,
		"manager_role_select", "Select League Manager roles", 0, 10)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	msgcomps := adminRoleSelect
	msgcomps = append(msgcomps, leagueMgrRoleSelect...)

	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Staff Role select",
			Description: `
**Admin Roles**
Select the roles that should have admin access.

***NOTE**
Users with the discord Administrator permission 
will have access regardless of the roles set here.*

**League Manager roles**
Select the roles that should have league manager access
`,

			Color: 0x00ff00, // Green color
		},
		Components: msgcomps,
	}
	return contents, nil
}

func getRoleSelect(
	ctx context.Context,
	tx db.SafeTX,
	rolePerm uint16,
	customID string,
	placeholder string,
	minRoles int,
	maxRoles int,
) ([]discordgo.MessageComponent, error) {
	roles, err := models.GetRoles(ctx, tx, rolePerm)
	if err != nil {
		return nil, errors.Wrap(err, "models.GetRoles")
	}
	var defaults []discordgo.SelectMenuDefaultValue
	for _, role := range roles {
		defaults = append(defaults, discordgo.SelectMenuDefaultValue{
			ID:   role,
			Type: discordgo.SelectMenuDefaultValueRole,
		})
	}
	msgcomps := components.ActionRow(components.RoleSelect(
		customID,
		placeholder,
		defaults,
		minRoles,
		maxRoles,
	))
	return msgcomps, nil
}
