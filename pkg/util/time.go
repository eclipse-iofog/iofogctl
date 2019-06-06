package util

import (
	"fmt"
	"time"
)

var timeFormat = time.UnixDate

func Now() string {
	return time.Now().Format(timeFormat)
}

func FromInt(sec int64) string {
	return time.Unix(sec, 0).Format(timeFormat)
}

func Elapsed(from, to string) (diff string, err error) {
	fromTime, err := time.Parse(timeFormat, from)
	if err != nil {
		return
	}
	diffTime := time.Now().Sub(fromTime)
	diff = formatDuration(diffTime)
	return
}

func formatDuration(duration time.Duration) string {
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
