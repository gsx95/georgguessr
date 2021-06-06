package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"math/rand"
)

type RoomWithPredefinedArea struct {
	pkg.Room
	Continent string `json:"continent"`
	Country   string `json:"country"`
	City      string `json:"city"`
}

type RoomWithPlaces struct {
	pkg.Room
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


	positions, err := RandomPositionByArea(continent, country, cities, room.Rounds)

	for i, pos := range positions  {
		room.GamesRounds = append(room.GamesRounds, pkg.GameRound{
			No: i,
			StartPosition: pkg.GeoPoint{
				Lat: pos.Lat(),
				Lon: pos.Lon(),
			},
			Scores: map[string]pkg.Guess{},
		})
	}

	return createRoom(&room.Room)
}

func CreateRoomWithCustomAreas(reqBody string) (string, error) {
	room := pkg.Room{}
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
		room.GamesRounds = append(room.GamesRounds, pkg.GameRound{
			No: i,
			StartPosition: pkg.GeoPoint{
				Lat: lat,
				Lon: lon,
			},
			Scores: map[string]pkg.Guess{},
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
		point, err := RandomPosForCity(&City{Name: place.Name, Country: place.Country})
		if err != nil {
			i--
			fmt.Println(err)
		}
		room.GamesRounds = append(room.GamesRounds, pkg.GameRound{
			No: i,
			StartPosition: pkg.GeoPoint{
				Lat: point.Lat(),
				Lon: point.Lon(),
			},
			Scores: map[string]pkg.Guess{},
		})
	}
	return createRoom(&room.Room)
}

func CreateRoomUnlimited(reqBody string) (string, error) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	for i := 0; i < room.Rounds; i++ {
		lat, lon := RandomPosition()
		room.GamesRounds = append(room.GamesRounds, pkg.GameRound{
			No: i,
			StartPosition: pkg.GeoPoint{
				Lat: lat,
				Lon: lon,
			},
			Scores: map[string]pkg.Guess{},
		})
	}
	return createRoom(&room)
}

func createRoom(room *pkg.Room) (string, error) {
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

func checkRoomAttributes(room *pkg.Room) (*pkg.Room, error) {
	if room.MaxPlayers == 0 {
		return nil, errors.New("zero players not possible")
	}

	id := pkg.RandomRoomID()
	for RoomExists(id) {
		id = pkg.RandomRoomID()
	}
	room.ID = id
	room.Players = []string{}
	room.Status = "waiting"

	return room, nil
}
