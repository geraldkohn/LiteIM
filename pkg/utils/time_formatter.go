package utils

import "time"

// Get the current timestamp by Second
func GetCurrentTimestampBySecond() int64 {
	return time.Now().Unix()
}

// Convert timestamp to time.Time type
func UnixSecondToTime(second int64) time.Time {
	return time.Unix(second, 0)
}

// Get the current timestamp by Nano
func GetCurrentTimestampByNano() int64 {
	return time.Now().UnixNano()
}

// Convert timestamp to time.Time type
func UnixNanoSecondToTime(nanoSecond int64) time.Time {
	return time.Unix(0, nanoSecond)
}

// // Get the current timestamp by Mill
func GetCurrentTimestampByMill() int64 {
	return time.Now().UnixNano() / 1e6
}
