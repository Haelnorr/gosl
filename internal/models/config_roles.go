package models

import (
	"context"
	"database/sql"
	"gosl/pkg/db"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
)

const (
	PermAdmin         uint16 = 1 // Admin permission
	PermLeagueManager uint16 = 2 // League Manager permission
)

// Add a permission to the provided role
func AddPermission(ctx context.Context, tx *db.SafeWTX, roleid string, perm uint16) error {
	query := `
INSERT INTO config_roles (role_id, permission) 
VALUES (?, ?) ON CONFLICT DO NOTHING;
`
	_, err := tx.Exec(ctx, query, roleid, perm)
	return err
}
func Remove(ctx context.Context, tx *db.SafeWTX, roleid string, perm uint16) error {
	query := `
DELETE FROM config_roles WHERE role_id = ? AND permission = ?;
`
	_, err := tx.Exec(ctx, query, roleid, perm)
	return err
}

// Check if the provided role has the provided permission
func MemberHasPermission(
	ctx context.Context,
	tx db.SafeTX,
	s *discordgo.Session,
	guildID string,
	member *discordgo.Member,
	permid uint16,
) (bool, error) {
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
	args := make([]any, len(member.Roles)+1)
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

// Get all roles with the provided permission
func GetRoles(
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
func SetRoles(
	ctx context.Context,
	tx *db.SafeWTX,
	roles []string,
	permid uint16,
) error {
	args := make([]any, 0, len(roles)+1)
	query := `DELETE FROM config_roles WHERE permission = ?`
	args = []any{permid}
	if len(roles) != 0 {
		query = `
        DELETE FROM config_roles WHERE permission = ?
        AND role_id NOT IN (` + strings.Repeat("?,", len(roles)-1) + `?);
        `
		for _, role := range roles {
			args = append(args, role)
			err := AddPermission(ctx, tx, role, permid)
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
