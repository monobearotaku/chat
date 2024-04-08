package timeconv

import "time"

func TimeToString(tm time.Time) string {
	return tm.Format(time.RFC3339)
}

func StringToTime(tmStr string) time.Time {
	time, _ := time.Parse(time.RFC3339, tmStr)
	return time
}
