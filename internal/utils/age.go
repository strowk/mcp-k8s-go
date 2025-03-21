package utils

import (
	"fmt"
	"time"
)

// FormatAge converts a duration to a human-readable age string
func FormatAge(duration time.Duration) string {
	if duration.Hours() < 1 {
		return duration.Round(time.Minute).String()
	}
	if duration.Hours() < 24 {
		return duration.Round(time.Hour).String()
	}
	days := int(duration.Hours() / 24)
	return FormatDays(days)
}

// FormatDays provides a concise representation of days
func FormatDays(days int) string {
	if days < 7 {
		return fmt.Sprintf("%dd", days)
	}
	if days < 30 {
		weeks := days / 7
		return fmt.Sprintf("%dw", weeks)
	}
	months := days / 30
	return fmt.Sprintf("%dmo", months)
}
