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

var selectManagerRoles = &bot.Message{
	Label:       "Select Manager Roles",
	Purpose:     models.MsgSelectManagerRoles,
	GetContents: selectManagerRolesContents,
}

// Get the message contents for the select manager roles component
func selectManagerRolesContents(
	ctx context.Context,
	b *bot.Bot,
) (bot.MessageContents, error) {
	b.Logger.Debug().Msg("Setting up select manager roles components")
	timeout, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		return nil, errors.Wrap(err, "b.conn.Begin")
	}
	defer tx.Rollback()
	b.Logger.Debug().Msg("Getting default values for select admin roles components")
	roles, err := models.GetRoles(ctx, tx, models.PermLeagueManager)
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
		b.Logger.Debug().Msg("Retrieving select manager roles components")
		return "",
			&discordgo.MessageEmbed{
				Title:       "Manager roles",
				Description: "Select the roles that should have manager access",
				Color:       0x00ff00, // Green color
			},
			components.RoleSelect(
				"manager_role_select",
				"Select manager roles",
				defaultValues,
				0,
				10,
			)
	}, nil
}

// Handle an interaction with the select manager roles component
func handleSelectManagerRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
) error {
	selectManagerRoles := b.Channels[models.ChannelAdmin].Messages[models.MsgSelectManagerRoles]
	selectManagerRoles.StartUpdate(false)
	roles := i.MessageComponentData().Values
	err := models.SetRoles(ctx, tx, roles, models.PermLeagueManager)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (manager)")
	}
	droles := i.MessageComponentData().Resolved.Roles
	msg := "League Manager roles updated to:\n"
	for _, role := range roles {
		msg = msg + " - " + droles[role].Name + "\n"
	}
	b.Log().UserEvent(i.Member, msg)
	b.Reply(msg, i)
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		errch := make(chan error)
		b.Logger.Debug().Msg("Updating manager roles select")
		go selectManagerRoles.Update(ctx, errch)
		if <-errch != nil {
			b.Logger.Warn().Err(err).
				Msg("Failed to update select admin roles message after interaction")
		}
		close(errch)
	}()
	return nil
}
