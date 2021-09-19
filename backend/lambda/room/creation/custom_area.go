package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/planar"
	"math/rand"
)

type CreatorCustomArea struct{}

func (cr *CreatorCustomArea) CreateRoom(reqBody string) (string, error) {
	defer pkg.LogDuration(pkg.Track())
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := positions{}

	for i := 0; i < room.Rounds + additionalCreationTries; i++ {
		area := room.Areas[rand.Intn(len(room.Areas))]
		lat, lon := cr.randomPositionInArea(area)
		positions.Pos = append(positions.Pos, newRoundPos(i, lat, lon))
	}

	streetViews, err := getStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room, *streetViews)

	return createRoom(&room)
}

func (cr *CreatorCustomArea) randomPositionInArea(area []pkg.GeoPoint) (lat, lon float64) {
	defer pkg.LogDuration(pkg.Track())
	polygon := cr.pointsToPolygon(area)
	pointValid := false
	var point orb.Point
	bound := polygon.Bound()
	for !pointValid {
		lat := pkg.GetRandomFloat(bound.Min.Lat(), bound.Max.Lat())
		lon := pkg.GetRandomFloat(bound.Min.Lon(), bound.Max.Lon())
		point = orb.Point{lon, lat}
		pointValid = planar.PolygonContains(polygon, point)
	}
	return point.Lat(), point.Lon()
}

func (cr *CreatorCustomArea) pointsToPolygon(points []pkg.GeoPoint) (polygon orb.Polygon) {
	defer pkg.LogDuration(pkg.Track())
	var ring orb.Ring
	ring = []orb.Point{}
	for _, point := range points {
		ring = append(ring, *orbPoint(point.Lat, point.Lon))
	}
	return []orb.Ring{ring}
}


