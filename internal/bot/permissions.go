package bot

import (
	"context"
	"database/sql"
	"gosl/pkg/db"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	permissionAdmin         uint16 = 1
	permissionLeagueManager uint16 = 2
)

func addPermission(ctx context.Context, tx *db.SafeWTX, roleid string, perm uint16) error {
	query := `
INSERT INTO config_roles (role_id, permission) 
VALUES (?, ?) ON CONFLICT DO NOTHING;
`
	_, err := tx.Exec(ctx, query, roleid, perm)
	return err
}
func removePermission(ctx context.Context, tx *db.SafeWTX, roleid string, perm uint16) error {
	query := `
DELETE FROM config_roles WHERE role_id = ? AND permission = ?;
`
	_, err := tx.Exec(ctx, query, roleid, perm)
	return err
}

func hasPermission(
	ctx context.Context,
	tx db.SafeTX,
	s *discordgo.Session,
	guildID string,
	user *discordgo.User,
	permid uint16,
) (bool, error) {
	member, err := s.GuildMember(guildID, user.ID)
	if err != nil {
		return false, errors.Wrap(err, "s.GuildMember")
	}
	admin := member.Permissions&discordgo.PermissionAdministrator != 0
	if admin {
		return true, nil
	}
	if len(member.Roles) == 0 {
		return false, nil
	}
	query := `
SELECT 1 FROM config_roles WHERE 
    permission = ? AND 
    role_id IN (` + strings.Repeat("?,", len(member.Roles)-1) + `? ) LIMIT 1;
`
	args := make([]interface{}, len(member.Roles)+1)
	args[0] = permid
	for i, roleID := range member.Roles {
		args[i+1] = roleID
	}
	var exists int
	row, err := tx.QueryRow(ctx, query, args...)
	if err != nil {
		return false, nil
	}
	err = row.Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

func getRolesWithPermission(
	ctx context.Context,
	tx db.SafeTX,
	permid uint16,
) ([]string, error) {
	query := `SELECT role_id FROM config_roles WHERE permission = ?;`
	rows, err := tx.Query(ctx, query, permid)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()
	var roles []string
	for rows.Next() {
		var role string
		err := rows.Scan(&role)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// Grants the permission to the provided roles and removes it from any roles
// not provided
func setRolesForPermission(
	ctx context.Context,
	tx *db.SafeWTX,
	roles []string,
	permid uint16,
) error {
	args := make([]interface{}, 0, len(roles)+1)
	query := `DELETE FROM config_roles WHERE permission = ?`
	args = []interface{}{permid}
	if len(roles) != 0 {
		query = `
        DELETE FROM config_roles WHERE permission = ?
        AND role_id NOT IN (` + strings.Repeat("?,", len(roles)-1) + `?);
        `
		args = append(args, permid)
		for _, role := range roles {
			args = append(args, role)
			err := addPermission(ctx, tx, role, permid)
			if err != nil {
				return errors.Wrap(err, "addPermission")
			}
		}
	}
	_, err := tx.Exec(ctx, query, args...)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}
