package helper

import "math/rand"
import "github.com/twinj/uuid"


func GetRandom(min, max int) int {
	if max == min {
		return min
	}
	return rand.Intn(max - min) + min
}

func GetRandomFloat(min, max float64) float64 {
	if max == min {
		return min
	}
	return rand.Float64() * (max - min) + min
}

func RandomUUID() string {
	u := uuid.NewV4()
	return u.String()
}
