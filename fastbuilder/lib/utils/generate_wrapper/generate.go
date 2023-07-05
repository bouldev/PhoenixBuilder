package generate_wrapper

import (
	"strings"

	"github.com/google/uuid"
)

func GenerateUUIDStringLower() string {
	return uuid.New().String()
}

func GenerateUUIDStringUpper() string {
	return strings.ToUpper(GenerateUUIDStringLower())
}
