package utils

import "time"

func ExponentialBackoff(attempts int) time.Duration {
	base := 1 * time.Second
	duration := time.Duration(float64(base) * float64(int(1)<<uint(attempts))) // 2^attempt
	return duration
}
