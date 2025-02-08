package utils

import (
	"crypto/sha256"
	"fmt"
)

func GenerateETag(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%s", hash)
}
