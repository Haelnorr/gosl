package bot

import (
	"context"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/rs/zerolog"
)

func handleSelectLogChannelInteraction(
	ctx context.Context,
	tx *db.SafeTX,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	selectedChannel := i.MessageComponentData().Values[0]
	err := setChannelPurpose(ctx, tx, selectedChannel, channelLog)
	if err != nil {
		return err
	}
	emphemeralMessage("Updated log channel to "+selectedChannel, logger, s, i)
	return nil
}
