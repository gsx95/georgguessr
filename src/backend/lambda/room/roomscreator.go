package main

import (
	"encoding/json"
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

func CreateRoomWithPredefinedArea(reqBody string) (string, error){
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	country := room.Country
	cities := room.City

	positions := Positions{}

	randomPositions, err := RandomPositionByArea(country, cities, room.Rounds + 10)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("No point could be generated: %v", err))
	}

	for i, pos := range randomPositions  {
		positions.Pos = append(positions.Pos, NewRoundPos(i, pos.Lat(), pos.Lon()))
	}

	streetViews, err := GetStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room.Room, *streetViews)

	return createRoom(&room.Room)
}

func CreateRoomWithCustomAreas(reqBody string) (string, error) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := Positions{}

	for i := 0; i < room.Rounds + 10; i++ {
		area := room.Areas[rand.Intn(len(room.Areas))]
		lat, lon := RandomPositionInArea(area)
		positions.Pos = append(positions.Pos, NewRoundPos(i, lat, lon))
	}

	streetViews, err := GetStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room, *streetViews)

	return createRoom(&room)
}

func CreateRoomWithPlaces(reqBody string) (string, error) {
	room := RoomWithPlaces{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := Positions{}
	createErrors := map[string]bool{}
	placeFeatures := make(map[string]*geojson.Feature, len(room.Places))

	for _, place := range room.Places {
		feature, err := getBestFittingGeoJSONFeature(place.Name, place.Country, place.Pos)
		if err != nil {
			fmt.Println(err)
		}
		placeFeatures[place.getID()] = feature
	}

	if len(placeFeatures) == 0{
		return "", pkg.InternalErr(fmt.Sprintf("could not generate features for places %v", room.Places))
	}

	for i := 0; i < room.Rounds + 10; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		point, err := RandomPosForCity(placeFeatures[place.getID()], place.Pos)
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
		return "", pkg.InternalErr(fmt.Sprintf("Got errors while creating room: %s", msgs))
	}

	streetViews, err := GetStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room.Room, *streetViews)
	return createRoom(&room.Room)
}

func CreateRoomUnlimited(reqBody string) (string, error) {
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := Positions{}

	for i := 0; i < room.Rounds + 10; i++ {
		lat, lon := RandomPosition()
		positions.Pos = append(positions.Pos, NewRoundPos(i, lat, lon))
	}

	streetViews, err := GetStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room, *streetViews)

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
	writeRoomToDB(*room)
	return room.ID, nil
}

func checkRoomAttributes(room *pkg.Room) (*pkg.Room, error) {
	if room.MaxPlayers == 0 {
		return nil, pkg.BadRequestErr("zero players not possible")
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
