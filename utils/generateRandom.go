package utils

import (
	"fmt"
	"math/rand"
)

func GenerateCode() string {
	code := rand.Intn(100000)
	return fmt.Sprintf("%05d", code)
}
