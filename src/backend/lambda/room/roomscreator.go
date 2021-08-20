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

	country := room.Country
	cities := room.City

	positions := Positions{}

	randomPositions, err := RandomPositionByArea(country, cities, room.Rounds + 10)

	for i, pos := range randomPositions  {
		positions.Pos = append(positions.Pos, NewRoundPos(i, pos.Lat(), pos.Lon()))
	}

	streetViews := GetStreetviewPositions(positions, room.Rounds)
	addStreetViewToRoom(&room.Room, streetViews)

	return createRoom(&room.Room)
}

func CreateRoomWithCustomAreas(reqBody string) (string, error) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	positions := Positions{}

	for i := 0; i < room.Rounds + 10; i++ {
		area := room.Areas[rand.Intn(len(room.Areas))]
		lat, lon, err := RandomPositionInArea(area)
		if err != nil {
			i--
			fmt.Println(err)
			continue
		}
		positions.Pos = append(positions.Pos, NewRoundPos(i, lat, lon))
	}

	streetViews := GetStreetviewPositions(positions, room.Rounds)
	addStreetViewToRoom(&room, streetViews)

	return createRoom(&room)
}

func CreateRoomWithPlaces(reqBody string) (string, error) {
	room := RoomWithPlaces{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	positions := Positions{}

	for i := 0; i < room.Rounds + 10; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		point, err := RandomPosForCity(&City{Name: place.Name, Country: place.Country})
		if err != nil {
			i--
			fmt.Println(err)
			continue
		}
		positions.Pos = append(positions.Pos, NewRoundPos(i, point.Lat(), point.Lon()))
	}

	streetViews := GetStreetviewPositions(positions, room.Rounds)
	addStreetViewToRoom(&room.Room, streetViews)
	return createRoom(&room.Room)
}

func CreateRoomUnlimited(reqBody string) (string, error) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err
	}

	positions := Positions{}

	for i := 0; i < room.Rounds + 10; i++ {
		lat, lon := RandomPosition()
		positions.Pos = append(positions.Pos, NewRoundPos(i, lat, lon))
	}

	streetViews := GetStreetviewPositions(positions, room.Rounds)
	addStreetViewToRoom(&room, streetViews)

	return createRoom(&room)
}

func addStreetViewToRoom(room *pkg.Room, streetViews StreetViewIDs) {
	for _, streetView := range streetViews.Panos {
		room.GamesRounds = append(room.GamesRounds, pkg.GameRound{
			No: streetView.Round,
			StartPosition: pkg.GeoPoint{
				Lat: streetView.Pos.Lat,
				Lon: streetView.Pos.Lng,
			},
			PanoID: streetView.PanoID,
			Scores: map[string]pkg.Guess{},
		})
	}
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
