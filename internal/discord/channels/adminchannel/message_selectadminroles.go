package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/discord/components"
	"gosl/internal/models"
	"gosl/pkg/db"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

var selectAdminRoles = &bot.Message{
	Label:       "Select Admin Roles",
	Purpose:     models.MsgSelectAdminRoles,
	GetContents: selectAdminRolesContents,
}

// Get the message contents for the select admin roles component
func selectAdminRolesContents(
	ctx context.Context,
	b *bot.Bot,
) (bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select admin roles components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select admin roles components")
	roles, err := models.GetRoles(ctx, tx, models.PermAdmin)
	if err != nil {
		return nil, errors.Wrap(err, "getRolesWithPermission")
	}
	tx.Commit()

	var defaultValues []discordgo.SelectMenuDefaultValue
	for _, role := range roles {
		defaultValues = append(defaultValues, discordgo.SelectMenuDefaultValue{
			ID:   role,
			Type: discordgo.SelectMenuDefaultValueRole,
		})
	}
	return func() (
		string,
		*discordgo.MessageEmbed,
		[]discordgo.MessageComponent,
	) {
		b.Logger.Debug().Msg("Retreiving select admin roles components")
		return "",
			&discordgo.MessageEmbed{
				Title: "Admin roles",
				Description: `
Select the roles that should have admin access.

**NOTE**
Users with the discord Administrator permission 
will have access regardless of the roles set here.`,
				Color: 0x00ff00, // Green color
			},
			components.RoleSelect(
				"admin_role_select",
				"Select admin roles",
				defaultValues,
				0,
				10,
			)
	}, nil
}

// Handle an interaction with the select admin roles component
func handleSelectAdminRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	selectAdminRoles := b.Channels[models.ChannelAdmin].Messages[models.MsgSelectAdminRoles]
	selectAdminRoles.StartUpdate(false)
	roles := i.MessageComponentData().Values
	err := models.SetRoles(ctx, tx, roles, models.PermAdmin)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (admin)")
	}
	droles := i.MessageComponentData().Resolved.Roles
	msg := "Admin roles updated to:\n"
	for _, role := range roles {
		msg = msg + " - " + droles[role].Name + "\n"
	}
	b.Log().UserEvent(i.Member, msg)
	b.Reply(msg, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		errch := make(chan error)
		b.Logger.Debug().Msg("Updating admin roles select")
		go selectAdminRoles.Update(ctx, errch)
		if <-errch != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select admin roles message after interaction")
		}
	}()
	return nil
}
