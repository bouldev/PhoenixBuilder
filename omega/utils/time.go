package utils

import "time"

func TimeToString(t time.Time) string {
	return t.Format("2006-01-02/15:04:05")
}

func StringToTime(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02/15:04:05  -0700 MST", timeStr)
}

func StringToTimeWithLocal(timeStr string) (time.Time, error) {
	return time.Parse("2006-01-02/15:04:05 -0700 MST", timeStr)
}
