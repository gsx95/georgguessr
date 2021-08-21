package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb/geojson"
	"math/rand"
)

type RoomWithPredefinedArea struct {
	pkg.Room
	Country   string `json:"country"`
	City      string `json:"city"`
}

type RoomWithPlaces struct {
	pkg.Room
	Places []Place `json:"places"`
}

type Place struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	Id		int
}

func CreateRoomWithPredefinedArea(reqBody string) (string, error, int) {
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err, 400
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

func CreateRoomWithCustomAreas(reqBody string) (string, error, int) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err, 400
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

func CreateRoomWithPlaces(reqBody string) (string, error, int) {
	room := RoomWithPlaces{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err, 400
	}

	positions := Positions{}

	createErrors := map[string]bool{}

	placeFeatures := make(map[int]*geojson.Feature, len(room.Places))

	for i, place := range room.Places {
		feature, err := getBestFittingGeoJSONFeature(place.Name, place.Country)
		if err != nil {
			return "", err, 500
		}
		place.Id = i
		placeFeatures[place.Id] = feature
	}

	for i := 0; i < room.Rounds + 10; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		point, err := RandomPosForCity(placeFeatures[place.Id])
		if err != nil {
			fmt.Println(err)
			createErrors[err.Error()] = true
			continue
		}
		positions.Pos = append(positions.Pos, NewRoundPos(i, point.Lat(), point.Lon()))
	}

	if len(positions.Pos) < room.Rounds {
		msgs := ""
		for errMsg := range createErrors {
			msgs += errMsg + ";"
		}
		return "", errors.New("Got errors while creating room: " + msgs), 500
	}

	streetViews := GetStreetviewPositions(positions, room.Rounds)
	addStreetViewToRoom(&room.Room, streetViews)
	return createRoom(&room.Room)
}

func CreateRoomUnlimited(reqBody string) (string, error, int) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", err, 400
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

func createRoom(room *pkg.Room) (string, error, int) {
	room, err, status := checkRoomAttributes(room)
	if err != nil {
		return "", err, status
	}
	err = writeRoomToDB(*room)
	if err != nil {
		return "", err, 500
	}
	return room.ID, nil, 200
}

func checkRoomAttributes(room *pkg.Room) (*pkg.Room, error, int) {
	if room.MaxPlayers == 0 {
		return nil, errors.New("zero players not possible"), 400
	}

	id := pkg.RandomRoomID()
	for RoomExists(id) {
		id = pkg.RandomRoomID()
	}
	room.ID = id
	room.Players = []string{}
	room.Status = "waiting"

	return room, nil, 200
}
