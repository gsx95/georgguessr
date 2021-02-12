package data

import (
	"encoding/json"
	"errors"
	"math/rand"
)

func CreateRoomUnlimted(reqBody string) (string, error) {
	id, err := createRoom(reqBody)
	if err != nil {
		return "", err
	}
	return id, nil
}


func createRoom(body string) (roomId string, err error) {

	room := Room{}
	err = json.Unmarshal([]byte(body), &room)
	if err != nil {
		return "", err
	}

	if room.Name == "" {
		return "", errors.New("no room name provided")
	}

	if room.MaxPlayers == 0 {
		return "", errors.New("zero players not possible")
	}

	id := generateRoomID()
	room.ID = id
	room.Players = 0
	room.Status = "waiting"
	err = CreateRoom(room)
	if err != nil {
		return "", err
	}
	return id, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRoomID() string {
	b := make([]byte, 80)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}