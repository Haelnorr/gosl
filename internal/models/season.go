package models

import (
	"context"
	"database/sql"
	"fmt"
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
	err = AddLeague(ctx, tx, s.ID, "Open")
	if err != nil {
		return nil, errors.Wrap(err, "AddLeague (Open)")
	}
	err = AddLeague(ctx, tx, s.ID, "IM")
	if err != nil {
		return nil, errors.Wrap(err, "AddLeague (IM)")
	}
	err = AddLeague(ctx, tx, s.ID, "Pro")
	if err != nil {
		return nil, errors.Wrap(err, "AddLeague (Pro)")
	}
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
		return nil, nil
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
			endTime = TimeInLocale(endTime, locale)
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
			finalsEndTime = TimeInLocale(finalsEndTime, locale)
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

func (s *Season) GetApprovedTeams(ctx context.Context, tx db.SafeTX) (*[]Team, error) {
	query := `
SELECT t.id, t.abbreviation, t.name, t.manager_id, p.name, t.color 
FROM team t 
JOIN team_registration tr ON tr.team_id = t.id
JOIN player p ON t.manager_id = p.id
WHERE tr.season_id = ? COLLATE NOCASE
AND tr.approved = 1 AND tr.placed = 0;`
	rows, err := tx.Query(ctx, query, s.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var teams []Team
	for rows.Next() {
		var team Team
		var color string
		err = rows.Scan(&team.ID, &team.Abbreviation, &team.Name,
			&team.ManagerID, &team.ManagerName, &color)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, nil
			}
			return nil, errors.Wrap(err, "rows.Scan")
		}
		colorint, err := hexToInt(color)
		if err != nil {
			team.Color = 0x181825
		} else {
			team.Color = colorint
		}
		teams = append(teams, team)
	}
	return &teams, nil
}

func (s *Season) GetApprovedFreeAgents(
	ctx context.Context,
	tx db.SafeTX,
) (*[]Player, error) {
	query := `
SELECT p.id, p.slap_id, p.name, p.discord_id 
FROM player p 
JOIN free_agent_registration fa ON fa.player_id = p.id
WHERE fa.season_id = ? COLLATE NOCASE
AND fa.approved = 1 AND fa.placed = 0;`
	rows, err := tx.Query(ctx, query, s.ID)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Query")
	}
	var players []Player
	for rows.Next() {
		var player Player
		err = rows.Scan(&player.ID, &player.SlapID, &player.Name, &player.DiscordID)
		if err != nil {
			return nil, errors.Wrap(err, "row.Scan")
		}
		players = append(players, player)
	}
	return &players, nil
}

func (s *Season) RegisterFreeAgent(
	ctx context.Context,
	tx *db.SafeWTX,
	playerID uint16,
	preferredLeague string,
) (*FreeAgentRegistration, error) {
	query := `
INSERT INTO free_agent_registration(player_id, season_id, preferred_league)
VALUES (?, ?, ?);
`
	if preferredLeague != "Open" && preferredLeague != "IM" && preferredLeague != "Pro" {
		return nil, errors.New("Invalid division, must be 'Open', 'IM', or 'Pro'")
	}
	res, err := tx.Exec(ctx, query, playerID, s.ID, preferredLeague)
	if err != nil {
		return nil, errors.Wrap(err, "tx.Exec")
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, errors.Wrap(err, "res.LastInsertId")
	}
	app, err := GetFreeAgentRegistration(ctx, tx, uint32(id))
	if err != nil {
		return nil, errors.Wrap(err, "GetFreeAgentRegistration")
	}
	return app, nil
}
