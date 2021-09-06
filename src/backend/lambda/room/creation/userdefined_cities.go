package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb/geojson"
	"math/rand"
)

type CreatorUserdefinedCities struct {}

type RoomWithPlaces struct {
	pkg.Room
	Places []place `json:"places"`
}

func (cr *CreatorUserdefinedCities) CreateRoom(reqBody string) (string, error) {
	room := RoomWithPlaces{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := positions{}
	createErrors := map[string]bool{}
	placeFeatures := make(map[string]*geojson.Feature, len(room.Places))

	for _, place := range room.Places {
		feature, err := getBestFittingGeoJSONFeature(place.Name, place.Country, place.Pos)
		if err != nil {
			fmt.Println(err)
		}
		placeFeatures[cr.getPlaceID(place)] = feature
	}

	if len(placeFeatures) == 0{
		return "", pkg.InternalErr(fmt.Sprintf("could not generate features for places %v", room.Places))
	}

	for i := 0; i < room.Rounds +additionalCreationTries; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		point, err := randomPosForCity(placeFeatures[cr.getPlaceID(place)], place.Pos)
		if err != nil {
			fmt.Println(err)
			createErrors[err.Error()] = true
			continue
		}
		positions.Pos = append(positions.Pos, newRoundPos(i, point.Lat(), point.Lon()))
	}

	if len(positions.Pos) < room.Rounds {
		msgs := ""
		for errMsg := range createErrors {
			msgs += errMsg + ";"
		}
		return "", pkg.InternalErr(fmt.Sprintf("Got errors while creating room: %s", msgs))
	}

	streetViews, err := getStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room.Room, *streetViews)
	return createRoom(&room.Room)
}

func (cr *CreatorUserdefinedCities) getPlaceID(p place) string {
	return fmt.Sprintf("%s_%s_%d", p.Name, p.Country, p.Pop)
}