package presenter

import (
	"fmt"
	"time"
)

// FormatTimeSince formats a time duration as a human-readable "X ago" string.
// Returns formats like "5 minutes ago", "2.5 hours ago", or "3 days ago".
// This is the verbose format suitable for detailed displays.
func FormatTimeSince(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Hour {
		return fmt.Sprintf("%.0f minutes ago", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.1f hours ago", duration.Hours())
	} else {
		return fmt.Sprintf("%.0f days ago", duration.Hours()/24)
	}
}

// FormatTimeSinceCompact formats a time duration as a compact "X ago" string.
// Returns formats like "5m ago", "2.5h ago", or "3d ago".
// This is the compact format suitable for table displays with limited space.
func FormatTimeSinceCompact(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Hour {
		return fmt.Sprintf("%.0fm ago", duration.Minutes())
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%.1fh ago", duration.Hours())
	} else {
		return fmt.Sprintf("%.0fd ago", duration.Hours()/24)
	}
}
