package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb/geojson"
	"log"
	"math/rand"
)

type CreatorSearchedPlaces struct {}

type RoomWithPlaces struct {
	pkg.Room
	Places []place `json:"places"`
}

func (cr *CreatorSearchedPlaces) CreateRoom(reqBody string) (roomId string, genPos *Positions, err error) {
	defer pkg.LogDuration(pkg.Track())
	room := RoomWithPlaces{}
	err = json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	var positions Positions
	createErrors := map[string]bool{}
	placeFeatures := make(map[string]*geojson.Feature, len(room.Places))

	for _, place := range room.Places {
		feature, err := getBestFittingGeoJSONFeature(place.Name, place.Country, place.Pos)
		if err != nil {
			log.Println(err)
		}
		placeFeatures[cr.getPlaceID(place)] = feature
	}

	if len(placeFeatures) == 0{
		return "", nil, pkg.InternalErr(fmt.Sprintf("could not generate features for places %v", room.Places))
	}

	for i := 0; i < room.Rounds + additionalCreationTries; i++ {
		place := room.Places[rand.Intn(len(room.Places))]
		point, err := randomPosForCity(placeFeatures[cr.getPlaceID(place)], place.Pos)
		if err != nil {
			log.Println(err)
			createErrors[err.Error()] = true
			continue
		}
		positions = append(positions, newPosition(point.Lat(), point.Lon()))
	}

	if len(positions) < room.Rounds {
		msgs := ""
		for errMsg := range createErrors {
			msgs += errMsg + ";"
		}
		return "", nil, pkg.InternalErr(fmt.Sprintf("Got errors while creating room: %s", msgs))
	}

	roomId, err = saveRoom(&room.Room)
	if err != nil {
		return "", nil, err
	}
	return roomId, &positions, nil
}

func (cr *CreatorSearchedPlaces) getPlaceID(p place) string {
	return fmt.Sprintf("%s_%s_%d", p.Name, p.Country, p.Pop)
}