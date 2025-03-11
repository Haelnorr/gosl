package managerchannel

import (
	"context"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
)

func updateMessages(
	ctx context.Context,
	b *util.Bot,
) []error {
	msgs := []*messages.ChannelMessage{
		selectSeason,
		createSeason,
		activeSeasonInfo,
	}
	msgerrors := messages.UpdateChannelMessages(ctx, b, msgs)
	return msgerrors
}
