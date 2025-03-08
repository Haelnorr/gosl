package managerchannel

import (
	"context"
	"github.com/pkg/errors"
	"gosl/internal/discord/messages"
	"gosl/internal/discord/util"
)

func updateMessages(
	ctx context.Context,
	channelID string,
	b *util.Bot,
) error {
	// Select current season
	b.Logger.Debug().Msg("Updating season select")
	err := messages.UpdateChannelMessage(
		ctx,
		b,
		selectSeasonComponents,
		messages.ManagerSelectSeason,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "messages.UpdateChannelMessage (select season)")
	}
	// Create new season
	b.Logger.Debug().Msg("Updating create season")
	err = messages.UpdateChannelMessage(
		ctx,
		b,
		createSeasonComponents,
		messages.ManagerCreateSeason,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "messages.UpdateChannelMessage (create season)")
	}
	// Show active season
	b.Logger.Debug().Msg("Updating active season")
	err = messages.UpdateChannelMessage(
		ctx,
		b,
		activeSeasonComponents,
		messages.ManagerActiveSeason,
		channelID,
	)
	if err != nil {
		return errors.Wrap(err, "messages.UpdateChannelMessage (create season)")
	}
	return nil
}
