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

func (cr *CreatorCustomArea) CreateRoom(reqBody string) (roomId string, genPos *Positions, err error) {
	defer pkg.LogDuration(pkg.Track())
	room := pkg.Room{}
	err = json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	var positions Positions

	for i := 0; i < room.Rounds + additionalCreationTries; i++ {
		area := room.Areas[rand.Intn(len(room.Areas))]
		position := cr.randomPositionInArea(area)
		positions = append(positions, position)
	}

	roomId, err = saveRoom(&room)
	if err != nil {
		return "", nil, err
	}
	return roomId, &positions, nil
}

func (cr *CreatorCustomArea) randomPositionInArea(area []pkg.GeoPoint) position {
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
	return newPosition(point.Lat(), point.Lon())
}

func (cr *CreatorCustomArea) pointsToPolygon(points []pkg.GeoPoint) (polygon orb.Polygon) {
	defer pkg.LogDuration(pkg.Track())
	var ring orb.Ring
	ring = []orb.Point{}
	for _, point := range points {
		ring = append(ring, *orbPoint(point.Lat, point.Lng))
	}
	return []orb.Ring{ring}
}


