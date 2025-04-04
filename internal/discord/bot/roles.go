package bot

import (
	"context"
	"gosl/internal/models"
	"gosl/pkg/db"
	"slices"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

func (b *Bot) AddRoleToUser(
	player *models.Player,
	roleID string,
) error {
	err := b.Session.GuildMemberRoleAdd(b.Config.DiscordGuildID,
		player.DiscordID, roleID)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildMemberRoleAdd")
	}
	return nil
}

func (b *Bot) RemoveRoleFromUser(
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
		return errors.New("Multiple roles for that purpose configured, cannot list users")
	}
	roleID := roleIDs[0]

	err = b.Session.GuildMemberRoleRemove(b.Config.DiscordGuildID,
		player.DiscordID, roleID)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildMemberRoleRemove")
	}
	return nil
}

func (b *Bot) CreateRole(name string, color *int, mentionable bool) (string, error) {
	role, err := b.Session.GuildRoleCreate(b.Config.DiscordGuildID, &discordgo.RoleParams{
		Name:        name,
		Color:       color,
		Mentionable: &mentionable,
	})
	if err != nil {
		return "", errors.Wrap(err, "b.Session.GuildRoleCreate")
	}
	return role.ID, nil
}

func (b *Bot) DeleteRole(roleID string) error {
	err := b.Session.GuildRoleDelete(b.Config.DiscordGuildID, roleID)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildRoleDelete")
	}
	return nil
}

func (b *Bot) FloatRoleUnder(roleID string, targetRoleID string) error {
	roles, err := b.Session.GuildRoles(b.Config.DiscordGuildID)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildRoles")
	}
	b.Logger.Trace().Interface("roles", roles).Msg("before sort")
	newroles := moveRole(roles, roleID, targetRoleID)
	b.Logger.Trace().Interface("roles", newroles).Msg("after sort")
	_, err = b.Session.GuildRoleReorder(b.Config.DiscordGuildID, newroles)
	if err != nil {
		return errors.Wrap(err, "b.Session.GuildRoleReorder")
	}
	return nil
}

func (b *Bot) CheckRoleExists(roleID string) (bool, error) {
	role, err := b.Session.GuildRoleEdit(b.Config.DiscordGuildID, roleID, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}
		return false, errors.Wrap(err, "b.Session.GuildRoleEdit")
	}
	if role != nil {
		return true, nil
	}
	return false, nil
}

func moveRole(roles []*discordgo.Role, moveID string, targetID string) []*discordgo.Role {
	sort.Slice(roles, func(i, j int) bool {
		return roles[i].Position < roles[j].Position
	})
	var moveIndex, targetIndex int
	var foundMove, foundTarget bool
	for i, role := range roles {
		if role.ID == moveID {
			moveIndex = i
			foundMove = true
		}
		if role.ID == targetID {
			targetIndex = i
			foundTarget = true
		}
		if foundMove && foundTarget {
			break
		}
	}
	if !foundMove || !foundTarget {
		return roles
	}
	moveItem := roles[moveIndex]
	roles = slices.Delete(roles, moveIndex, moveIndex+1)
	if moveIndex < targetIndex {
		targetIndex--
	}
	newroles := slices.Insert(roles, targetIndex, moveItem)
	for i, role := range newroles {
		role.Position = i
	}
	return newroles
}
