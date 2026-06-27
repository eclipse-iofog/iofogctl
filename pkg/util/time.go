package util

import (
	"fmt"
	"time"
)

func NowUTC() string {
	return time.Now().Format(time.UnixDate)
}

func NowRFC() string {
	return time.Now().Format(time.RFC3339)
}

func FromIntUTC(sec int64) string {
	return time.Unix(sec, 0).Format(time.UnixDate)
}

func ElapsedUTC(from, to string) (diff string, err error) {
	fromTime, err := time.Parse(time.UnixDate, from)
	if err != nil {
		return
	}
	diffTime := time.Since(fromTime)
	diff = FormatDuration(diffTime)
	return
}

func ElapsedRFC(from, to string) (diff string, err error) {
	fromTime, err := time.Parse(time.RFC3339, from)
	if err != nil {
		return
	}
	diffTime := time.Since(fromTime)
	diff = FormatDuration(diffTime)
	return
}

func FormatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	// Get days
	days := duration / (time.Hour * 24)
	duration -= days * (time.Hour * 24)

	// Get hours
	hours := duration / time.Hour
	duration -= hours * time.Hour

	// Get Minutes
	mins := duration / time.Minute
	duration -= mins * time.Minute

	// Get Seconds
	secs := duration / time.Second

	if days > 0 {
		return fmt.Sprintf("%dd%dh", days, hours)
	}

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, mins)
	}

	if mins > 0 {
		return fmt.Sprintf("%dm%ds", mins, secs)
	}

	return fmt.Sprintf("%ds", secs)
}
