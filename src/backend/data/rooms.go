package data

import (
	"backend/helper"
	"encoding/json"
	"errors"
)

type RoomWithPredefinedArea struct {
	Room
	Continent       string `json:"continent"`
	Country       string `json:"country"`
	City       string `json:"city"`

}

func CreateRoomWithPredefinedArea(reqBody string) (string, error) {
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}
	return createRoom(&room.Room)
}

func CreateRoomUnlimited(reqBody string) (string, error) {
	room := Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	for i := 0; i < room.Rounds; i++ {
		lat, lon := RandomPosition()
		room.GamesRounds = append(room.GamesRounds, GameRound{
			No: i,
			StartPosition: GeoPoint{
				Lat: lat,
				Lon: lon,
			},
		})
	}
	return createRoom(&room)
}

func createRoom(room *Room) (string, error) {
	room, err := checkRoomAttributes(room)
	if err != nil {
		return "", err
	}
	err = writeRoomToDB(*room)
	if err != nil {
		return "", err
	}
	return room.ID, nil
}

func checkRoomAttributes(room *Room) (*Room, error) {
	if room.Name == "" {
		return nil, errors.New("no room name provided")
	}

	if room.MaxPlayers == 0 {
		return nil, errors.New("zero players not possible")
	}

	id := helper.RandomUUID()
	room.ID = id
	room.Players = 0
	room.Status = "waiting"

	return room, nil
}