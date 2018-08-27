package util

import "strconv"

// Str2float64 converts a string to float64
func Str2float64(str string) float64 {
	value, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return -1
	}
	return value
}
