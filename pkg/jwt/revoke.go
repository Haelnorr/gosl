package jwt

import (
	"context"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Revoke a token by adding it to the database
func RevokeToken(ctx context.Context, tx *db.SafeWTX, t Token) error {
	jti := t.GetJTI()
	exp := t.GetEXP()
	query := `INSERT INTO jwtblacklist (jti, exp) VALUES (?, ?)`
	_, err := tx.Exec(ctx, query, jti, exp)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

// Check if a token has been revoked. Returns true if not revoked.
func CheckTokenNotRevoked(ctx context.Context, tx db.SafeTX, t Token) (bool, error) {
	jti := t.GetJTI()
	query := `SELECT 1 FROM jwtblacklist WHERE jti = ? LIMIT 1`
	rows, err := tx.Query(ctx, query, jti)
	if err != nil {
		return false, errors.Wrap(err, "tx.Query")
	}
	defer rows.Close()
	revoked := rows.Next()
	return !revoked, nil
}
