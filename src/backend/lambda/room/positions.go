package main

import (
	_ "embed"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/knakk/sparql"
	"github.com/mmcloughlin/spherand"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geo"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const getCityBoundariesUrl = "https://nominatim.openstreetmap.org/search.php?q=%s+%s&polygon_geojson=1&format=geojson"
const wikiDataUrl = "https://query.wikidata.org/sparql"

var minPopulation = map[string]int{
	"pop_gt_100k": 100000,
	"pop_gt_500k": 500000,
	"pop_gt_1kk":  1000000,
	"pop_gt_5kk":  5000000,
}

type CountryName struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Place struct {
	Name    string `json:"name"`
	Country string `json:"country"`
	ID      string
	Pop     int
	Pos Position `json:"location"`
}

func (p *Place) getID() string {
	return fmt.Sprintf("%s_%s_%d", p.Name, p.Country, p.Pop)
}

func NewCity(name, country string, pop int, pos Position) *Place {
	return &Place{
		Pop: pop,
		Name: name,
		Country: country,
		ID: fmt.Sprintf("%s_%s_%d", name, country, pop),
		Pos: pos,
	}
}

//go:embed sparql/getCountryByCode.query
var getCountryByCode string
//go:embed sparql/getCitiesByCountryAndPop.query
var getCityByCountryAndPop string
//go:embed sparql/getRandomCityForPop.query
var getRandomCityForPop string

func RandomPosition() (lat, lon float64) {
	lat, lon = spherand.Geographical()
	return
}


func RandomPositionInArea(area []pkg.GeoPoint) (lat, lon float64) {
	polygon := pointsToPolygon(area)
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

// Returns error if no valid point could be generated
func RandomPositionByArea(country string, cities string, count int) (positions []*orb.Point, err error) {

	minPop := minPopulation[cities]
	if country != "all" {
		countryData, err := queryWikiData(fmt.Sprintf(getCountryByCode, strings.ToUpper(country)))
		if err != nil {
			return nil, err
		}
		countryEntity := countryData.Results.Bindings[0]["country"].Value
		entityParts := strings.Split(countryEntity, "/")
		countryID := entityParts[len(entityParts)-1]

		query := fmt.Sprintf(getCityByCountryAndPop, countryID, minPop, time.Now().String())

		results, err := queryWikiData(query)
		if err != nil {
			return nil, err
		}
		var citiesSlice []*Place
		bindings := results.Results.Bindings
		for _, b := range bindings {
			cityPop, err := strconv.Atoi(b["maxPopulation"].Value)
			cityName := b["cityLabel"].Value
			locationString := b["location"].Value
			if err != nil {
				log.Fatal(err)
			}
			if cityPop > minPop {
				pos, err := wikiDataStringToPos(locationString)
				if err != nil {
					return nil, err
				}
				newCity := NewCity(cityName, country, cityPop, *pos)
				citiesSlice = append(citiesSlice, newCity)
			}
		}
		return randomPosForCities(citiesSlice, count)
	}

	query := fmt.Sprintf(getRandomCityForPop, minPop, time.Now().String())
	results, err := queryWikiData(query)
	if err != nil {
		return nil, err
	}
	res := results.Results.Bindings
	var possibleCities []*Place
	for _, result := range res {
		cityName := result["cityLabel"].Value
		countryName := result["countryLabel"].Value
		locationString := result["location"].Value
		pop, err := strconv.Atoi(result["population"].Value)
		if err != nil {
			fmt.Printf("error while trying to convert wikiData pop %s to int\n", result["population"].Value)
			continue
		}
		pos, err := wikiDataStringToPos(locationString)
		if err != nil {
			fmt.Println(err)
			continue
		}
		possibleCities = append(possibleCities, NewCity(cityName, countryName, pop, *pos))
	}
	return randomPosForCities(possibleCities, count)
}

// Returns error if no valid point could be generated
func RandomPosForCity(feature *geojson.Feature, originalPlace Position) (point *orb.Point, err error) {

	pointValid := false
	box := feature.BBox.Bound()

	for !pointValid {
		lat := pkg.GetRandomFloat(box.Min.Lat(), box.Max.Lat())
		lon := pkg.GetRandomFloat(box.Min.Lon(), box.Max.Lon())
		point = &orb.Point{lon, lat}
		pointValid, err = isPointInsidePolygon(feature, point, &originalPlace)
		if err != nil {
			fmt.Println(err)
		}
	}

	return point, nil
}

func getBestFittingGeoJSONFeature(city, country string, position Position) (*geojson.Feature, error) {
	fmt.Println(fmt.Sprintf(getCityBoundariesUrl, city, country))
	req, err := http.NewRequest("GET", fmt.Sprintf(getCityBoundariesUrl, city, country), nil)
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

	featureCollection, err := geojson.UnmarshalFeatureCollection(bodyBytes)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Error unmarshalling geojson response from openstreemmaps: %v %v", string(bodyBytes), err))
	}


	var filteredPlaces []*geojson.Feature

	// get all features where the place lies within its boundaries
	for _, feature := range featureCollection.Features {
		if ok, _ := isPointInsidePolygon(feature, orbPoint(position.Lat, position.Lng), nil); ok {
			filteredPlaces = append(filteredPlaces, feature)
		}
	}

	if len(filteredPlaces) == 0 {
		fmt.Println("No feature has polygon in which the place lies")
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


func isPointInsidePolygon(feature *geojson.Feature, point *orb.Point, originalPlace *Position) (bool, error) {
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

func pointsToPolygon(points []pkg.GeoPoint) (polygon orb.Polygon) {
	var ring orb.Ring
	ring = []orb.Point{}
	for _, point := range points {
		ring = append(ring, *orbPoint(point.Lat, point.Lon))
	}
	return []orb.Ring{ring}
}

func randomPosForCities(possibleCities []*Place, count int) (positions []*orb.Point, err error) {

	var cities []*Place
	cityFeatures := make(map[string]*geojson.Feature, 0)

	for i := 0; i < count; i++ {
		randomCity := possibleCities[pkg.GetRandom(0, len(possibleCities)-1)]
		cities = append(cities, randomCity)
	}

	for _, city := range cities {
		if _, exists := cityFeatures[city.ID]; !exists {
			feature, err := getBestFittingGeoJSONFeature(city.Name, city.Country, city.Pos)
			if err != nil {
				fmt.Println(err)
				continue
			}
			cityFeatures[city.ID] = feature
		}
	}

	if len(cityFeatures) == 0 {
		return nil, pkg.InternalErr(fmt.Sprintf("Could not get feature for any provided place %v", cities))
	}

	for i := range cities {
		point, err := RandomPosForCity(cityFeatures[cities[i].ID], cities[i].Pos)
		if err != nil {
			fmt.Println(err)
			i--
		}
		positions = append(positions, point)
	}

	if len(positions) != count {
		return nil, pkg.InternalErr(fmt.Sprintf("could not create %d rounds, created %d", count, len(positions)))
	}
	return
}

func queryWikiData(query string) (*sparql.Results, error) {
	fmt.Println("Wikidata query: " + query)
	repo, err := sparql.NewRepo(wikiDataUrl)
	if err != nil {
		return nil, pkg.InternalErr(err.Error())
	}
	results, err := repo.Query(query)
	if err != nil {
		return nil, pkg.InternalErr(err.Error())
	}
	fmt.Printf("Wikidata results: %v\n", results.Results.Bindings)
	return results, nil
}

func orbPoint(Lat, Lng float64) *orb.Point {
	p := &orb.Point{}
	p[0] = Lng
	p[1] = Lat
	return p
}

func wikiDataStringToPos(queryResult string) (*Position, error) {      // e.g.  Point(7.099722222 50.733888888)
	coordsTxt := strings.Trim(queryResult, "Point()")
	coords := strings.Split(coordsTxt, " ")

	lng, err := strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Could not parse longitude %s from position %s", lng, coordsTxt))
	}
	lat, err := strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Could not parse longitude %s from position %s", lng, coordsTxt))
	}

	return &Position{
		Lng: lng,
		Lat: lat,
	}, nil

}