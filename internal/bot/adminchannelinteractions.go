package bot

import (
	"context"
	"gosl/pkg/db"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func handleSelectLogChannelInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	selectedChannel := i.MessageComponentData().Values[0]
	err := setChannelPurpose(ctx, tx, selectedChannel, channelLog)
	if err != nil {
		return errors.Wrap(err, "setChannelPurpose (log channel)")
	}
	emphemeralMessage("Updated log channel to "+selectedChannel, logger, s, i)
	return nil
}

func handleSelectAdminRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	roles := i.MessageComponentData().Values
	err := setRolesForPermission(ctx, tx, roles, permissionAdmin)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (admin)")
	}
	emphemeralMessage("updated admin roles", logger, s, i)
	return nil
}

func handleSelectManagerRolesInteraction(
	ctx context.Context,
	tx *db.SafeWTX,
	logger *zerolog.Logger,
	s *discordgo.Session,
	i *discordgo.InteractionCreate,
) error {
	roles := i.MessageComponentData().Values
	err := setRolesForPermission(ctx, tx, roles, permissionLeagueManager)
	if err != nil {
		return errors.Wrap(err, "setRolesForPermission (manager)")
	}
	emphemeralMessage("updated league manager roles", logger, s, i)
	return nil
}
