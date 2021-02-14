package helper

import "math/rand"
import "github.com/twinj/uuid"


func GetRandom(min, max int) int {
	return rand.Intn(max - min) + min
}

func RandomUUID() string {
	u := uuid.NewV4()
	return u.String()
}
