package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectTeamMgrRoles = &bot.Message{
	Label:       "Select Team Manager Roles",
	Purpose:     models.MsgSelectTeamMgrRoles,
	GetContents: selectTeamMgrRolesContents,
}

// Get the message contents for the select roles message
func selectTeamMgrRolesContents(
	ctx context.Context,
	b *bot.Bot,
) (*bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select team manager roles message")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// HACK:
	// we use a WTX here to force it to block until commit if an update to
	// the contents resulted in this function being called.
	// this is because the message update uses the Message.GetContents field
	// and doesnt take in a tx as an input
	// to fix this, either do something like accept a WG and wait until told to proceed
	// OR make sure all calls to update contents are preceeded by a commit
	tx, err := b.Conn.Begin(timeout, "selectTeamMgrRolesContents()")
	if err != nil {
		return nil, errors.Wrap(err, "b.Conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select roles components")
	proTeamManagerRoleSelect, err := getRoleSelect(ctx, tx, models.PermProTeamManager,
		"pro_team_manager_role_select", "Select Pro Team Manager role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	imTeamManagerRoleSelect, err := getRoleSelect(ctx, tx, models.PermIMTeamManager,
		"im_team_manager_role_select", "Select IM Team Manager role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}
	openTeamManagerRoleSelect, err := getRoleSelect(ctx, tx, models.PermOpenTeamManager,
		"open_team_manager_role_select", "Select Open Team Manager role", 1, 1)
	if err != nil {
		return nil, errors.Wrap(err, "getRoleSelect")
	}

	tx.Commit()
	msgcomps := []discordgo.MessageComponent{}
	msgcomps = append(msgcomps, proTeamManagerRoleSelect...)
	msgcomps = append(msgcomps, imTeamManagerRoleSelect...)
	msgcomps = append(msgcomps, openTeamManagerRoleSelect...)

	contents := &bot.MessageContents{
		Embed: &discordgo.MessageEmbed{
			Title: "Team Manager Role select",
			Description: `
**Team Manager roles**
Select the roles that should be used for team managers
`,

			Color: 0x00ff00, // Green color
		},
		Components: msgcomps,
	}
	return contents, nil
}
