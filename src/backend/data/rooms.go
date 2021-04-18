package data

import (
	"backend/helper"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
)

type RoomWithPredefinedArea struct {
	Room
	Continent string `json:"continent"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

type RoomWithPlaces struct {
	Room
	Places [] struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"places"`
}

func CreateRoomWithPredefinedArea(reqBody string) (string, error) {
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	continent := room.Continent
	country := room.Country
	cities := room.City

	for i := 0; i < room.Rounds; i++ {
		lat, lon, err := RandomPositionByArea(continent, country, cities)
		if err != nil {
			i--
			fmt.Println(err)
		}
		room.GamesRounds = append(room.GamesRounds, GameRound{
			No: i,
			StartPosition: GeoPoint{
				Lat: lat,
				Lon: lon,
			},
			Scores: map[string]Guess{},
		})
	}

	return createRoom(&room.Room)
}

func CreateRoomWithCustomAreas(reqBody string) (string, error) {
	room := Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	for i := 0; i < room.Rounds; i++ {
		area := room.Areas[rand.Intn(len(room.Areas))]
		lat, lon, err := RandomPositionInArea(area)
		if err != nil {
			i--
			fmt.Println(err)
		}
		room.GamesRounds = append(room.GamesRounds, GameRound{
			No: i,
			StartPosition: GeoPoint{
				Lat: lat,
				Lon: lon,
			},
			Scores: map[string]Guess{},
		})
	}
	return createRoom(&room)
}

func CreateRoomWithPlaces(reqBody string) (string, error) {
	room := RoomWithPlaces{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	for i := 0; i < room.Rounds; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		lat, lon, err := RandomPosForCity(place.Name, place.Country)
		if err != nil {
			i--
			fmt.Println(err)
		}
		room.GamesRounds = append(room.GamesRounds, GameRound{
			No: i,
			StartPosition: GeoPoint{
				Lat: lat,
				Lon: lon,
			},
			Scores: map[string]Guess{},
		})
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
			Scores: map[string]Guess{},
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
	if room.MaxPlayers == 0 {
		return nil, errors.New("zero players not possible")
	}

	id := helper.RandomRoomID()
	for RoomExists(id) {
		id = helper.RandomRoomID()
	}
	room.ID = id
	room.Players = []string{}
	room.Status = "waiting"

	return room, nil
}
