package adminchannel

import (
	"context"
	"github.com/pkg/errors"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
)

// Update the messages in the admin channel
func updateMessages(
	ctx context.Context,
	channelID string,
	b *util.Bot,
) error {
	// Select log channel
	b.Logger.Debug().Msg("Updating log channel select")
	err := messages.UpdateChannelMessage(
		ctx,
		b,
		selectLogChannelContents,
		messages.AdminSelectLogChannel,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectLogChannel)")
	}

	// Select admin roles
	b.Logger.Debug().Msg("Updating admin roles select")
	err = messages.UpdateChannelMessage(
		ctx,
		b,
		selectAdminRolesContents,
		messages.AdminSelectAdminRoles,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectAdminRoles)")
	}

	// Select manager roles
	b.Logger.Debug().Msg("Updating manager roles select")
	err = messages.UpdateChannelMessage(
		ctx,
		b,
		selectManagerRolesContents,
		messages.AdminSelectManagerRoles,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "updateChannelMessage (selectManagerRoles)")
	}
	return nil
}
