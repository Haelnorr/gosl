package bot

import (
	"fmt"
	"time"
)

// Format the time as a discord time stamp with type F (Full Date & Time)
//
// Friday, March 14, 2025 7:35 PM
func DiscordDateTime(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:F>"
	return fmt.Sprintf(str, unix)
}

// Format the time as a discord time stamp with type D (Date)
//
// March 14, 2025
func DiscordDate(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:D>"
	return fmt.Sprintf(str, unix)
}

// Format the time as a discord time stamp with type t (Time)
//
// 7:35 PM
func DiscordTime(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:t>"
	return fmt.Sprintf(str, unix)
}

// Format the time as a discord time stamp with type R (time until)
//
// in 7 days
func DiscordUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	unix := t.Unix()
	str := "<t:%v:R>"
	return fmt.Sprintf(str, unix)
}

// Format the time as a discord time stamp with type F and type R
// (Full Date & Time and time until)
//
// Friday, March 14, 2025 7:35 PM (in 7 days)
func DiscordDateTimeUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	datetime := DiscordDateTime(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", datetime, until)
}

// Format the time as a discord time stamp with type D and type R (Date and time until)
//
// March 14, 2025 (in 7 days)
func DiscordDateUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	date := DiscordDate(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", date, until)
}

// Format the time as a discord time stamp with type t and type R (Time and time until)
//
// 7:35 PM
func DiscordTimeUntil(t *time.Time) string {
	if t == nil {
		return "Not set"
	}
	time := DiscordTime(t)
	until := DiscordUntil(t)
	return fmt.Sprintf("%s (%s)", time, until)
}
