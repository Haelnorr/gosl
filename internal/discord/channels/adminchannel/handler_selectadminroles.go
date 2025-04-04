package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle an interaction with the select admin roles component
func handleSelectAdminRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	msgSelectRoles, err := b.GetMessage(models.ChannelAdmin, models.MsgSelectStaffRoles)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	if !msgSelectRoles.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	roles := i.MessageComponentData().Values
	err = models.SetRoles(ctx, tx, roles, models.PermAdmin)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (admin)")
	}
	droles := i.MessageComponentData().Resolved.Roles
	msg := "**Admin roles updated to:**  \n"
	for _, role := range roles {
		msg = msg + " - " + droles[role].Name + "\n"
	}
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	errch := make(chan error)
	b.Logger.Debug().Msg("Updating admin roles select")
	go msgSelectRoles.Update(ctx, tx, errch)
	hadErr := false
	for err := range errch {
		if err != nil {
			hadErr = true
			msg := "Failed to update message after interaction"
			b.DoubleError(msg, err)
		}
	}
	if hadErr {
		return errors.New("Failed to update message(s) after interaction")
	}
	return nil
}
