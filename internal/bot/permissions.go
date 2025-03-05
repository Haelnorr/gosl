package bot

import (
	"context"
	"database/sql"
	"gosl/pkg/db"
	"strings"

	"github.com/bwmarrin/discordgo"
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
		return false, err
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
