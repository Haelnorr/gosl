package adminchannel

import (
	"context"
	"gosl/internal/discord/channels/channels"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	adminChannelName string = "gosl-bot-admin"
)

// Setup the admin channel
func Setup(
	wg *sync.WaitGroup,
	errch chan error,
	ctx context.Context,
	b *util.Bot,
) {
	defer wg.Done()
	b.Logger.Debug().Msg("Setting up admin channel")
	timeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	tx, err := b.Conn.Begin(timeout)
	if err != nil {
		errch <- errors.Wrap(err, "conn.Begin")
		return
	}
	defer tx.Rollback()
	channelID, err := findExistingAdminChannel(ctx, tx, b)
	if err != nil {
		errch <- errors.Wrap(err, "findExistingAdminChannel")
		return
	}
	if channelID == "" {
		channelID, err = createNewAdminChannel(ctx, tx, b)
		if err != nil {
			errch <- errors.Wrap(err, "createNewLogChannel")
			return
		}
		if !channels.CheckExists(channelID, b.Session) {
			errch <- errors.New("Unknown error occurred setting up log channel")
			return
		}
	}
	tx.Commit()
	b.Logger.Info().Str("channel_id", channelID).Msg("Admin channel is ready")

	err = updateMessages(ctx, channelID, b)
	if err != nil {
		errch <- errors.Wrap(err, "b.updateAdminMessages")
		return
	}
	b.Session.AddHandler(handleInteractions(ctx, b))
	b.Logger.Info().Msg("Admin channel setup complete")
}

func findExistingAdminChannel(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
) (string, error) {
	b.Logger.Debug().Msg("Getting admin channel ids")
	channelIDs, err := channels.GetChannels(ctx, tx, channels.PurposeAdmin)
	if err != nil {
		return "", errors.Wrap(err, "channels.GetChannels")
	}
	var adminChannelID string
	for _, channelID := range channelIDs {
		if exists := channels.CheckExists(channelID, b.Session); exists {
			b.Logger.Debug().Msg("Admin channel found")
			adminChannelID = channelID
		} else {
			b.Logger.Debug().Msg("Removing dead channel ID from database")
			channels.RemovePurpose(ctx, tx, channelID, channels.PurposeAdmin)
		}
	}
	return adminChannelID, nil
}

func createNewAdminChannel(
	ctx context.Context,
	tx *db.SafeWTX,
	b *util.Bot,
) (string, error) {
	b.Logger.Debug().Msg("Creating new admin channel")
	channel, err := b.Session.GuildChannelCreate(
		b.GuildID, adminChannelName, discordgo.ChannelTypeGuildText)
	if err != nil {
		return "", errors.Wrap(err, "s.GuildChannelCreate")
	}

	b.Logger.Debug().Msg("Adding new channel to database")
	err = channels.AddPurpose(ctx, tx, channel.ID, channels.PurposeAdmin)
	if err != nil {
		return "", errors.Wrap(err, "channels.AddPurpose")
	}
	return channel.ID, nil
}
