package creation

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"log"
)

//go:embed resources/continents.geojson
var continentsGeoJson string

type RoomWithPredefinedArea struct {
	pkg.Room
	Continent string `json:"continent"`
	Country string `json:"country"`
}

type CreatorCountryContinent struct{}

const countryBoundaryApiEndpoint = "https://nominatim.openstreetmap.org/search.php?country=%s&polygon_geojson=1&format=geojson"

func (cr *CreatorCountryContinent) CreateRoom(reqBody string) (string, error) {
	defer pkg.LogDuration(pkg.Track())
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	continentCode := room.Continent
	countryCode := room.Country
	var selectedFeature *geojson.Feature
	log.Println("get boundaries")

	if countryCode == "all" {
		selectedFeature, err = getContinentBoundaries(continentCode)
	} else {
		selectedFeature, err = getCountryBoundary(countryCode)
		// if the country boundary is a multipolygon, only use the polygon with the biggest area
		// for example, france consists of multiple polygons spread around the world
		// the bounds are basically the whole world, so trying to generate a random point inside the bounds
		// that lie inside the multipolygon is very difficult and usually takes more than 30 seconds
		if multiPolygon, isMulti := selectedFeature.Geometry.(orb.MultiPolygon); isMulti {
			biggestPolygon, newBound := cr.getBiggestPolygon(multiPolygon)
			selectedFeature.Geometry = biggestPolygon
			selectedFeature.BBox = geojson.NewBBox(newBound)
		}
	}

	log.Println("got boundaries, generate random position")
	randomPositions := cr.randomPositionsInFeature(selectedFeature, room.Rounds+additionalCreationTries)
	positions := positions{}
	for i, pos := range randomPositions {
		positions.Pos = append(positions.Pos, newRoundPos(i, pos.Lat(), pos.Lon()))
	}
	log.Println("generate streetviews")

	streetViews, err := getStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room.Room, *streetViews)

	return createRoom(&room.Room)
}


func getContinentBoundaries(continentCode string) (*geojson.Feature, error) {
	continents, err := geojson.UnmarshalFeatureCollection([]byte(continentsGeoJson))
	if err != nil {
		return nil, pkg.InternalErr("Could not unmarshal continents geojson")
	}
	for _, continentFeature := range continents.Features {
		if continentFeature.Properties.MustString("CONTINENT") == continentCode {
			return continentFeature, nil
		}
	}
	return nil, pkg.InternalErr("Could not get continent boundaries for continent " + continentCode)
}

func getCountryBoundary(countryCode string) (feature *geojson.Feature, err error) {
	featureCollection, err := requestGeoJson(fmt.Sprintf(countryBoundaryApiEndpoint, countryCode))
	if err != nil {
		return nil, err
	}
	if len(featureCollection.Features) > 1 {
		log.Println("there was more then one feature for country boundary")
	}
	return featureCollection.Features[0], nil
}

func (cr *CreatorCountryContinent) randomPositionsInFeature(feature *geojson.Feature, amount int) []*orb.Point {
	var points []*orb.Point
	for i := 0; i < amount; i++ {
		point := cr.randomPosInFeature(feature)
		points = append(points, &point)
	}
	return points
}

func (cr *CreatorCountryContinent) randomPosInFeature(feature *geojson.Feature) orb.Point {
	bound := feature.BBox.Bound()
	for {
		lat := pkg.GetRandomFloat(bound.Min.Lat(), bound.Max.Lat())
		lon := pkg.GetRandomFloat(bound.Min.Lon(), bound.Max.Lon())
		point := orb.Point{lon, lat}
		log.Printf("try generated point %v inside bounds %v \n", point, bound)
		if multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon); isMulti {
			if planar.MultiPolygonContains(multiPoly, point) {
				return point
			}
		}
		if poly, isPoly := feature.Geometry.(orb.Polygon); isPoly {
			if planar.PolygonContains(poly, point) {
				return point
			}
		}
	}
}

func (cr *CreatorCountryContinent) getBiggestPolygon(multiPolygon orb.MultiPolygon) (biggestPolygon orb.Polygon, bound orb.Bound) {
	biggestPolygonArea := float64(0)
	for _, polygon := range multiPolygon {
		area := planar.Area(polygon)
		if area > biggestPolygonArea {
			biggestPolygon = polygon
			biggestPolygonArea = area
		}
	}

	minLat := float64(1000)
	minLon := float64(1000)
	maxLat := float64(-1000)
	maxLon := float64(-1000)

	for _, coords := range biggestPolygon[0] {
		if coords.Lat() > maxLat {
			maxLat = coords.Lat()
		}
		if coords.Lat() < minLat {
			minLat = coords.Lat()
		}
		if coords.Lon() > maxLon {
			maxLon = coords.Lon()
		}
		if coords.Lon() < minLon {
			minLon = coords.Lon()
		}
	}

	min := orb.Point{minLon, minLat}
	max := orb.Point{maxLon, maxLat}

	newBound := orb.Bound{
		Max:max,
		Min: min,
	}
	return biggestPolygon, newBound
}
