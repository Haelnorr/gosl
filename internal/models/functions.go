package models

import (
	"strconv"
	"strings"
	"time"
)

// Set the timezone for the time to the configured locale
// Preserves the time of day. E.g. if the time provided has timestamp at midnight
// the returned time will be at midnight in the specified locale
func TimeInLocale(t *time.Time, locale string) *time.Time {
	loc, _ := time.LoadLocation(locale)
	dateInTZ := t.In(loc)
	return &dateInTZ
}

// Converts a string value in format "2006-01-02T15:04:05Z07:00" to time.Time.
// If nil value provided, will return nil
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

// If u is 0, returns false, else returns true
func uint16ToBool(u uint16) bool {
	if u == 0 {
		return false
	} else {
		return true
	}
}

// Parses string in format "02/01/2006" to time.Time
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

// Converts a time.Time to format "2006-01-02T15:04:05Z07:00". Time is nil returns ""
func formatISO8601(t *time.Time) string {
	if t == nil {
		return ""
	}
	formatted := t.Format(time.RFC3339)
	return formatted
}

// Formats time.Time to format "02/01/2006"
func DateStr(t *time.Time) string {
	if t == nil {
		return ""
	}
	formatted := t.Format("02/01/2006")
	return formatted
}

// Parses a hex string to an integer. E.g. color hex codes #00FF00 -> 65280
func hexToInt(hexStr string) (int, error) {
	if strings.HasPrefix(hexStr, "#") {
		hexStr = hexStr[1:]
	}

	value, err := strconv.ParseInt(hexStr, 16, 0)
	if err != nil {
		return 0, err
	}

	return int(value), nil
}
