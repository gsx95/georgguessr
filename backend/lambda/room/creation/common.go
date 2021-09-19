package creation

import (
	"fmt"
	"georgguessr.com/lambda-room/db"
	"georgguessr.com/pkg"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Positions []position

type position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type place struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	ID      string
	Pop     int
	Pos     position `json:"location"`
}

const additionalCreationTries = 30
const getCityBoundariesUrl = "https://nominatim.openstreetmap.org/search.php?q=%s+%s&polygon_geojson=1&format=geojson"


func newPosition(lat, lng float64) position {
	return position{
			Lat: lat,
			Lng: lng,
	}
}

func saveRoom(room *pkg.Room) (string, error) {
	defer pkg.LogDuration(pkg.Track())
	if room.MaxPlayers == 0 {
		return "", pkg.BadRequestErr("zero players not possible")
	}

	id := pkg.RandomRoomID()
	for db.RoomExists(id) {
		id = pkg.RandomRoomID()
	}
	room.ID = id
	room.Players = []string{}

	db.WriteRoomToDB(*room)
	return room.ID, nil
}

func orbPoint(Lat, Lng float64) *orb.Point {
	p := &orb.Point{}
	p[0] = Lng
	p[1] = Lat
	return p
}


// Returns error if no valid point could be generated
func randomPosForCity(feature *geojson.Feature, originalPlace position) (point *orb.Point, err error) {
	defer pkg.LogDuration(pkg.Track())
	pointValid := false
	box := feature.BBox.Bound()

	for !pointValid {
		lat := pkg.GetRandomFloat(box.Min.Lat(), box.Max.Lat())
		lon := pkg.GetRandomFloat(box.Min.Lon(), box.Max.Lon())
		point = &orb.Point{lon, lat}
		pointValid, err = isPointInsidePolygon(feature, point, &originalPlace)
		if err != nil {
			log.Println(err)
		}
	}

	return point, nil
}

func requestGeoJson(url string) (*geojson.FeatureCollection, error) {
	log.Println("calling " + url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Error getting geojson featuers from openstreemmaps: %v", err))
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Error getting geojson featuers from openstreemmaps: %v", err))
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Error reading geojson response from openstreemmaps: %v", err))
	}
	log.Println("unmarshal geojson body")

	featureCollection, err := geojson.UnmarshalFeatureCollection(bodyBytes)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Error unmarshalling geojson response from openstreemmaps: %v %v", string(bodyBytes), err))
	}
	return featureCollection, nil
}

func getBestFittingGeoJSONFeature(city, country string, position position) (*geojson.Feature, error) {
	defer pkg.LogDuration(pkg.Track())

	featureCollection, err := requestGeoJson(fmt.Sprintf(getCityBoundariesUrl, city, country))
	if err != nil {
		return nil, err
	}

	var filteredPlaces []*geojson.Feature

	// get all features where the place lies within its boundaries
	for _, feature := range featureCollection.Features {
		if ok, _ := isPointInsidePolygon(feature, orbPoint(position.Lat, position.Lng), nil); ok {
			filteredPlaces = append(filteredPlaces, feature)
		}
	}

	if len(filteredPlaces) == 0 {
		log.Println("No feature has polygon in which the place lies")
		filteredPlaces = featureCollection.Features
	}

	filteredPlacesCopy := filteredPlaces
	filteredPlaces = []*geojson.Feature{}

	for _, feature := range filteredPlacesCopy {
		featureCategory := strings.ToLower(feature.Properties.MustString("category"))
		if featureCategory == "place" || featureCategory == "boundary" {
			filteredPlaces = append(filteredPlaces, feature)
		}
	}

	// if there is no "place", use previous filtered places
	if len(filteredPlaces) == 0 {
		filteredPlaces = filteredPlacesCopy
	}

	sort.Slice(filteredPlaces, func(f1, f2 int) bool {
		imp1 := filteredPlaces[f1].Properties.MustFloat64("importance", 0)
		imp2 := filteredPlaces[f2].Properties.MustFloat64("importance", 0)
		return imp1 > imp2
	})

	return filteredPlaces[0], nil
}


func isPointInsidePolygon(feature *geojson.Feature, point *orb.Point, originalPlace *position) (bool, error) {
	defer pkg.LogDuration(pkg.Track())
	if multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon); isMulti {
		// if its a multipolygon, we only care for the polygon in which the originalPlace lies
		if originalPlace != nil {
			orig := *orbPoint(originalPlace.Lat, originalPlace.Lng)
			for _, pol := range multiPoly {
				if planar.PolygonContains(pol, orig) && planar.PolygonContains(pol, *point) {
					feature.Geometry = pol
					return true, nil
				}
			}
			return false, nil
		}

		// no polygon contains the original place (or no original place given) - just use the bigger polygon
		polyLen := 0
		var polygon orb.Polygon
		for _, pol := range multiPoly {
			if len(pol[0]) >= polyLen {
				polyLen = len(pol[0])
				polygon = pol
			}
		}
		feature.Geometry = polygon
		return planar.PolygonContains(polygon, *point), nil
	}

	// if its one polygon, check if the generated point is inside of it
	if polygon, isPolygon := feature.Geometry.(orb.Polygon); isPolygon {
		return planar.PolygonContains(polygon, *point), nil
	}

	// if its a point, check if the generated point lies within 5.000 meters
	if p, isPoint := feature.Geometry.(orb.Point); isPoint {
		return geo.Distance(p, *point) < 5000, nil
	}

	return false, pkg.InternalErr("geometry of feature is neither multipolygon nor polygon nor point")
}