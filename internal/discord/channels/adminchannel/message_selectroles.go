package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectRoles = &bot.Message{
	Label:       "Select Admin Roles",
	Purpose:     models.MsgSelectRoles,
	GetContents: selectRolesContents,
}

// Get the message contents for the select roles message
func selectRolesContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select roles message")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select roles components")
	adminroles, err := models.GetRoles(ctx, tx, models.PermAdmin)
	if err != nil {
		return nil, errors.Wrap(err, "getRolesWithPermission")
	}
	managerroles, err := models.GetRoles(ctx, tx, models.PermLeagueManager)
	if err != nil {
		return nil, errors.Wrap(err, "getRolesWithPermission")
	}
	tx.Commit()

	var adminRoleDefaults []discordgo.SelectMenuDefaultValue
	for _, role := range adminroles {
		adminRoleDefaults = append(adminRoleDefaults, discordgo.SelectMenuDefaultValue{
			ID:   role,
			Type: discordgo.SelectMenuDefaultValueRole,
		})
	}
	var managerRoleDefaults []discordgo.SelectMenuDefaultValue
	for _, role := range managerroles {
		managerRoleDefaults = append(managerRoleDefaults, discordgo.SelectMenuDefaultValue{
			ID:   role,
			Type: discordgo.SelectMenuDefaultValueRole,
		})
	}
	msgcomps := components.RoleSelect(
		"admin_role_select",
		"Select admin roles",
		adminRoleDefaults,
		0,
		10,
	)
	msgcomps = append(msgcomps, components.RoleSelect(
		"manager_role_select",
		"Select League manager roles",
		managerRoleDefaults,
		0,
		10,
	)...)
	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Role select",
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
