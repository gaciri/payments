package utils

import "time"

func FmtTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
