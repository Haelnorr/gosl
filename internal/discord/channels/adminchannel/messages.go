package adminchannel

import (
	"context"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
)

// Update the messages in the admin channel
func updateMessages(
	ctx context.Context,
	b *util.Bot,
) []error {
	msgs := []*messages.ChannelMessage{
		selectLogChannel,
		selectAdminRoles,
		selectManagerRoles,
	}
	msgerrors := messages.UpdateChannelMessages(ctx, b, msgs)
	return msgerrors
}
