package models

import (
	"time"
)

func parseISO8601(isostr *string) *time.Time {
	if isostr == nil {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *isostr)
	if err != nil {
		return nil
	}
	return &parsed
}

func uint16ToBool(u uint16) bool {
	if u == 0 {
		return false
	} else {
		return true
	}
}

func parseTextDate(datestr string) *time.Time {
	format := "02/01/2006"
	if datestr == "" {
		return nil
	}
	parsed, err := time.Parse(format, datestr)
	if err != nil {
		return nil
	}
	return &parsed
}

func formatISO8601(t *time.Time) string {
	if t == nil {
		return ""
	}
	formatted := t.Format(time.RFC3339)
	return formatted
}

func DateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	formatted := t.Format("02/01/2006")
	return formatted
}
