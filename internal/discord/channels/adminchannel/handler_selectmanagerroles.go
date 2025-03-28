package adminchannel

import (
	"context"
	"gosl/internal/discord/bot"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

// Handle an interaction with the select manager roles component
func handleSelectManagerRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	b *bot.Bot,
	i *discordgo.InteractionCreate,
	ack *bool,
) error {
	b.Acknowledge(i, ack)
	msgSelectRoles, err := b.GetMessage(models.ChannelAdmin, models.MsgSelectRoles)
	if err != nil {
		return errors.Wrap(err, "b.GetMessage")
	}
	if !msgSelectRoles.StartUpdate(false) {
		b.SlowDown(i, *ack)
		return nil
	}
	roles := i.MessageComponentData().Values
	err = models.SetRoles(ctx, tx, roles, models.PermLeagueManager)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (manager)")
	}
	droles := i.MessageComponentData().Resolved.Roles
	msg := "**League Manager roles updated to:**  \n"
	for _, role := range roles {
		msg = msg + " - " + droles[role].Name + "\n"
	}
	b.Log().UserEvent(i.Member, msg)
	err = b.FollowUp(msg, i)
	if err != nil {
		return errors.Wrap(err, "b.FollowUp")
	}
	// Spin off updating the message so it doesnt block/get blocked by the transaction
	// and runs as soon as the interaction is completed
	go func() {
		errch := make(chan error)
		b.Logger.Debug().Msg("Updating manager roles select")
		go msgSelectRoles.Update(ctx, errch)
		for err := range errch {
			if err != nil {
				msg := "Failed to update message after interaction"
				b.DoubleError(msg, err)
			}
		}
	}()
	return nil
}
