package utils

import "github.com/google/uuid"

func GenerateUID() string {
	return uuid.NewString()
}
