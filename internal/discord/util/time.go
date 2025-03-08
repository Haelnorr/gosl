package util

import (
	"fmt"
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

func DiscordDateTime(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:F>"
	return fmt.Sprintf(str, unix)
}
func DiscordDate(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:D>"
	return fmt.Sprintf(str, unix)
}
func DiscordTime(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:t>"
	return fmt.Sprintf(str, unix)
}
func DiscordUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:R>"
	return fmt.Sprintf(str, unix)
}
func DiscordDateTimeUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	datetime := DiscordDateTime(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", datetime, until)
}
func DiscordDateUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	date := DiscordDate(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", date, until)
}
func DiscordTimeUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	time := DiscordTime(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", time, until)
}
