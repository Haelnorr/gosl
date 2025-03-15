package models

import (
	"context"
	"gosl/pkg/db"

	"github.com/pkg/errors"
)

// Model of the league table in the database
// Each record in this table covers a single league for a given season
type League struct {
	ID       uint16 // unique ID
	Division string // division of the league i.e. Open, IM, Pro
	SeasonID string // FK -> Season.ID
}

func AddLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	division string,
) error {
	if division != "Open" && division != "IM" && division != "Pro" {
		return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	query := `
INSERT INTO league (division, season_id, enabled)
VALUES (?, ?, 1)
ON CONFLICT (division, season_id)
DO UPDATE SET enabled = 1;
`
	_, err := tx.Exec(ctx, query, division, seasonID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func GetLeagues(
	ctx context.Context,
	tx db.SafeTX,
	seasonID string,
) (*[]League, error) {
	query := `SELECT id, division FROM league WHERE season_id = ? AND enabled = 1;`
	rows, err := tx.Query(ctx, query, seasonID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var leagues []League
	for rows.Next() {
		var league League
		err = rows.Scan(&league.ID, &league.Division)
		if err != nil {
			return nil, errors.Wrap(err, "rows.Scan")
		}
		league.SeasonID = seasonID
		leagues = append(leagues, league)
	}
	return &leagues, nil
}

func RemoveLeague(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	division string,
) error {
	if division != "Open" && division != "IM" && division != "Pro" {
		return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	query := `UPDATE league SET enabled = 0 WHERE division = ? AND season_id = ?;`
	_, err := tx.Exec(ctx, query, division, seasonID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	return nil
}

func SetLeagues(
	ctx context.Context,
	tx *db.SafeWTX,
	seasonID string,
	divisions []string,
) error {
	allDivs := map[string]bool{
		"Open": false,
		"IM":   false,
		"Pro":  false,
	}
	for _, division := range divisions {
		if division != "Open" && division != "IM" && division != "Pro" {
			return errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
		}
		allDivs[division] = true
	}
	for division, enabled := range allDivs {
		if enabled {
			err := AddLeague(ctx, tx, seasonID, division)
			if err != nil {
				return errors.Wrap(err, "AddLeague")
			}
		} else {
			err := RemoveLeague(ctx, tx, seasonID, division)
			if err != nil {
				return errors.Wrap(err, "RemoveLeague")
			}
		}
	}
	return nil
}
