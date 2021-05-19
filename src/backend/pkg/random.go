package pkg

import "math/rand"

var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func GetRandom(min, max int) int {
	if max == min {
		return min
	}
	return rand.Intn(max-min) + min
}

func GetRandomFloat(min, max float64) float64 {
	if max == min {
		return min
	}
	return rand.Float64()*(max-min) + min
}

func RandomRoomID() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
