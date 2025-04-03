package bot

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

func (b *Bot) AddRoleToUser(
	ctx context.Context,
	tx db.SafeTX,
	player *models.Player,
	perm uint16,
) error {
	roleIDs, err := models.GetRoles(ctx, tx, perm)
	if err != nil {
		return errors.Wrap(err, "models.GetRoles")
	}
	if len(roleIDs) == 0 {
		return errors.New("No roles for that purpose configured")
	}
	if len(roleIDs) > 1 {
		return errors.New("Multiple roles for that purpose configured, cannot add to user")
	}
	roleID := roleIDs[0]

	err = b.Session.GuildMemberRoleAdd(b.Config.DiscordGuildID,
		player.DiscordID, roleID)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildMemberRoleAdd")
	}
	return nil
}
