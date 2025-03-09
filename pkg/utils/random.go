package utils

import (
	"math/rand"
	"time"
)

const (
	// CharSet for generating random codes
	letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomCode generates a random string of specified length
func GenerateRandomCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

// GenerateRandomID generates a random ID
func GenerateRandomID() string {
	return GenerateRandomCode(16)
}
