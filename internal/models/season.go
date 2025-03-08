package models

import (
	"context"
	"database/sql"
	"fmt"
	"gosl/internal/discord/util"
	"gosl/pkg/db"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// Model of the season table in the database
// Each record in the table represents a single season covering all leagues
type Season struct {
	ID               string     // unique identifier e.g. S21
	Name             string     // unique display name e.g. Season 21
	Start            *time.Time // timestamp ISO8601 format
	RegSeasonEnd     *time.Time // timestamp ISO8601 format
	FinalsEnd        *time.Time // timestamp ISO8601 format
	Active           bool       // is season active
	RegistrationOpen bool       // is team registration open
}

func CreateSeason(
	ctx context.Context,
	tx *db.SafeWTX,
	id string,
	name string,
) (*Season, error) {
	query := `INSERT INTO season (id, name) VALUES (?, ?);`
	_, err := tx.Exec(ctx, query, id, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			msg := ""
			if strings.Contains(err.Error(), "season.name") {
				msg = fmt.Sprintf("Season name must be unique: '%s' is taken", name)
			} else if strings.Contains(err.Error(), "season.id") {
				msg = fmt.Sprintf("Season ID must be unique: '%s' is taken", id)
			}
			return nil, errors.New(msg)
		}
		return nil, errors.Wrap(err, "tx.Exec")
	}
	var s Season
	s.ID = id
	s.Name = name
	s.Active = false
	s.RegistrationOpen = false
	return &s, nil
}

func GetSeason(ctx context.Context, tx db.SafeTX, id string) (*Season, error) {
	query := `
SELECT id, name, start, reg_season_end, finals_end, active, registration_open
FROM season WHERE id = ?;
`
	row, err := tx.QueryRow(ctx, query, id)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	season, err := scanSeason(row)
	if err != nil {
		return nil, errors.Wrap(err, "scanSeason")
	}
	return season, nil
}

func GetSeasons(ctx context.Context, tx db.SafeTX) ([]*Season, error) {
	var seasons []*Season
	query := `
SELECT id, name, start, reg_season_end, finals_end, active, registration_open
FROM season;
`
	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	for rows.Next() {
		season, err := scanSeason(rows)
		if err != nil {
			return nil, errors.Wrap(err, "scanSeason")
		}
		seasons = append(seasons, season)
	}
	return seasons, nil
}

func GetActiveSeason(ctx context.Context, tx db.SafeTX) (*Season, error) {
	query := `
SELECT id, name, start, reg_season_end, finals_end, active, registration_open
FROM season WHERE active = 1;
`
	row, err := tx.QueryRow(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "tx.QueryRow")
	}
	season, err := scanSeason(row)
	if err == sql.ErrNoRows {
		return &Season{
			ID:               "NOACTIVESEASON",
			Name:             "No active season",
			Active:           false,
			RegistrationOpen: false,
		}, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "scanSeason")
	}
	return season, nil
}

// Sets the active season to the given season ID. Providing "NOACTIVESEASON" as the ID will
// set any active season as inactive. DB has a trigger setup to ensure only 1
// season is set as active at once
func SetActiveSeason(ctx context.Context, tx *db.SafeWTX, id string) error {
	if id == "NOACTIVESEASON" {
		query := `UPDATE season SET active = 0 WHERE active = 1;`
		_, err := tx.Exec(ctx, query)
		if err != nil {
			return errors.Wrap(err, "tx.Exec")
		}
	} else {
		query := `UPDATE season SET active = 1 WHERE id = ?`
		_, err := tx.Exec(ctx, query, id)
		if err != nil {
			return errors.Wrap(err, "tx.Exec")
		}
	}
	return nil
}

func scanSeason(row any) (*Season, error) {
	var s Season
	var start *string
	var regSeasonEnd *string
	var finalsEnd *string
	var active uint16
	var regOpen uint16
	switch r := row.(type) {
	case *sql.Row:
		err := r.Scan(&s.ID, &s.Name, &start, &regSeasonEnd, &finalsEnd, &active, &regOpen)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, err
			}
			return nil, errors.Wrap(err, "row.Scan")
		}
	case *sql.Rows:
		err := r.Scan(&s.ID, &s.Name, &start, &regSeasonEnd, &finalsEnd, &active, &regOpen)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, err
			}
			return nil, errors.Wrap(err, "row.Scan")
		}
	default:
		return nil, errors.New("invalid row type")
	}
	startParsed := parseISO8601(start)
	s.Start = startParsed
	regSeasonEndParsed := parseISO8601(regSeasonEnd)
	s.RegSeasonEnd = regSeasonEndParsed
	finalsEndParsed := parseISO8601(finalsEnd)
	s.FinalsEnd = finalsEndParsed
	activeBool := uint16ToBool(active)
	regOpenBool := uint16ToBool(regOpen)
	s.Active = activeBool
	s.RegistrationOpen = regOpenBool

	return &s, nil
}

// Toggles the value of registration_open for the given season in both the database
// AND the Season struct. NOTE that if a rollback occurs before the transaction
// is committed, the Season struct will not be reverted to the previous value
func (s *Season) ToggleRegistration(ctx context.Context, tx *db.SafeWTX) error {
	query := `UPDATE season SET registration_open = ? WHERE id = ?;`
	openVal := 1
	if s.RegistrationOpen {
		openVal = 0
	}
	_, err := tx.Exec(ctx, query, openVal, s.ID)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	s.RegistrationOpen = !s.RegistrationOpen
	return nil
}

func (s *Season) RegistrationStatusString() string {
	if s.RegistrationOpen {
		return "Open"
	} else {
		return "Closed"
	}
}

func (s *Season) SetDates(
	ctx context.Context,
	tx *db.SafeWTX,
	startStr, endStr, finalsStr string,
	locale string,
) error {
	query := ""
	var startTime *time.Time
	var endTime *time.Time
	var finalsEndTime *time.Time
	if startStr == "" {
		query = query + `UPDATE season SET start = NULL WHERE id = "` + s.ID + `";`
	} else {
		startTime = parseTextDate(startStr)
		if startTime != nil {
			startTime = util.TimeInLocale(startTime, locale)
			timeStr := formatISO8601(startTime)
			newq := fmt.Sprintf(
				`UPDATE season SET start = "%s" WHERE id = "%s";`,
				timeStr,
				s.ID,
			)
			query = query + newq
		}
	}
	if endStr == "" {
		query = query + `UPDATE season SET reg_season_end = NULL WHERE id = "` + s.ID + `";`
	} else {
		endTime = parseTextDate(endStr)
		if endTime != nil {
			endTime = util.TimeInLocale(endTime, locale)
			timeStr := formatISO8601(endTime)
			newq := fmt.Sprintf(
				`UPDATE season SET reg_season_end = "%s" WHERE id = "%s";`,
				timeStr,
				s.ID,
			)
			query = query + newq
		}
	}
	if finalsStr == "" {
		query = query + `UPDATE season SET finals_end = NULL WHERE id = "` + s.ID + `";`
	} else {
		finalsEndTime = parseTextDate(finalsStr)
		if finalsEndTime != nil {
			finalsEndTime = util.TimeInLocale(finalsEndTime, locale)
			timeStr := formatISO8601(finalsEndTime)
			newq := fmt.Sprintf(
				`UPDATE season SET finals_end = "%s" WHERE id = "%s";`,
				timeStr,
				s.ID,
			)
			query = query + newq
		}
	}
	_, err := tx.Exec(ctx, query)
	if err != nil {
		return errors.Wrap(err, "tx.Exec")
	}
	s.Start = startTime
	s.RegSeasonEnd = endTime
	s.FinalsEnd = finalsEndTime
	return nil
}
