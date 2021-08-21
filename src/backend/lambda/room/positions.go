package main

import (
	_ "embed"
	"errors"
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

type City struct {
	ID      string
	Name    string
	Pop     int
	Country string
}

func NewCity(name, country string, pop int) *City {
	return &City{
		Pop: pop,
		Name: name,
		Country: country,
		ID: fmt.Sprintf("%s_%s_%d", name, country, pop),
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

func RandomPositionInArea(area []pkg.GeoPoint) (lat, lon float64, err error) {
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
	return point.Lat(), point.Lon(), nil
}

func RandomPositionByArea(country string, cities string, count int) (positions []*orb.Point, err error) {

	if country != "all" {
		return getRandomPosByCountryAndPop(minPopulation[cities], country, count)
	}
	return getRandomPosByPop(minPopulation[cities], count)
}

func RandomPosForCity(feature *geojson.Feature) (point *orb.Point, err error) {

	pointValid := false
	box := feature.BBox.Bound()

	for !pointValid {
		lat := pkg.GetRandomFloat(box.Min.Lat(), box.Max.Lat())
		lon := pkg.GetRandomFloat(box.Min.Lon(), box.Max.Lon())
		point = &orb.Point{lon, lat}
		pointValid = isPointInsidePolygon(feature, point)
	}

	return point, nil
}

func getBestFittingGeoJSONFeature(city, country string) (*geojson.Feature, error) {
	fmt.Println(fmt.Sprintf(getCityBoundariesUrl, city, country))
	req, err := http.NewRequest("GET", fmt.Sprintf(getCityBoundariesUrl, city, country), nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	featureCollection, err := geojson.UnmarshalFeatureCollection(bodyBytes)
	if err != nil {
		return nil, err
	}

	var placeFeatures []*geojson.Feature

	for _, feature := range featureCollection.Features {
		featureCategory := strings.ToLower(feature.Properties.MustString("category"))
		if featureCategory == "place" {
			placeFeatures = append(placeFeatures, feature)
		}
	}

	// if there is no "place", use all types of nodes as fallback
	if len(placeFeatures) == 0 {
		placeFeatures = featureCollection.Features
	}

	sort.Slice(placeFeatures, func(f1, f2 int) bool {
		imp1 := placeFeatures[f1].Properties.MustFloat64("importance", 0)
		imp2 := placeFeatures[f2].Properties.MustFloat64("importance", 0)
		return imp1 > imp2
	})

	return placeFeatures[0], nil
}

func isPointInsidePolygon(feature *geojson.Feature, point *orb.Point) bool {

	// if its a polygon, we only care for the polygon with more boundary nodes
	if multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon); isMulti {
		polyLen := 0
		var polygon orb.Polygon
		for _, pol := range multiPoly {
			if len(pol[0]) >= polyLen {
				polyLen = len(pol[0])
				polygon = pol
			}
		}
		return planar.PolygonContains(polygon, *point)
	}

	// if its one polygon, check if the generated point is inside of it
	if polygon, isPolygon := feature.Geometry.(orb.Polygon); isPolygon {
		return planar.PolygonContains(polygon, *point)
	}

	// if its a point, check if the generated point lies within 5.000 meters
	if p, isPoint := feature.Geometry.(orb.Point); isPoint {
		return geo.Distance(p, *point) < 5000
	}

	panic("geometry of feature is neither multipolygon nor polygon nor point")
}

func pointsToPolygon(points []pkg.GeoPoint) (polygon orb.Polygon) {
	var ring orb.Ring
	ring = []orb.Point{}
	for _, point := range points {
		p := orb.Point{}
		p[0] = point.Lon
		p[1] = point.Lat
		ring = append(ring, p)
	}
	return []orb.Ring{ring}
}

func getRandomPosByCountryAndPop(minPop int, country string, count int) (positions []*orb.Point, err error) {
	fmt.Printf("get by country and pop: %s, %d\n", country, minPop)

	countryData := queryWikiData(fmt.Sprintf(getCountryByCode, strings.ToUpper(country)))
	countryEntity := countryData.Results.Bindings[0]["country"].Value
	entityParts := strings.Split(countryEntity, "/")
	countryID := entityParts[len(entityParts)-1]

	query := fmt.Sprintf(getCityByCountryAndPop, countryID, minPop, time.Now().String())

	results := queryWikiData(query)
	var cities []*City
	bindings := results.Results.Bindings
	for _, b := range bindings {
		cityPop, err := strconv.Atoi(b["maxPopulation"].Value)
		if err != nil {
			log.Fatal(err)
		}
		if cityPop > minPop {
			cities = append(cities, NewCity(b["cityLabel"].Value, country, cityPop))
		}
	}

	return randomPosForCities(cities, count)
}

func getRandomPosByPop(minPop int, count int) (positions []*orb.Point, err error) {
	query := fmt.Sprintf(getRandomCityForPop,
		minPop, time.Now().String())

	results := queryWikiData(query)

	res := results.Results.Bindings
	var possibleCities []*City
	for _, result := range res {
		cityName := result["cityLabel"].Value
		countryName := result["countryLabel"].Value
		pop, err := strconv.Atoi(result["population"].Value)
		if err != nil {
			log.Fatal(err)
		}
		possibleCities = append(possibleCities, NewCity(cityName, countryName, pop))
	}
	return randomPosForCities(possibleCities, count)
}

func randomPosForCities(possibleCities []*City, count int) (positions []*orb.Point, err error) {

	var cities []*City
	cityFeatures := make(map[string]*geojson.Feature, 0)

	for i := 0; i < count; i++ {
		randomCity := possibleCities[pkg.GetRandom(0, len(possibleCities)-1)]
		cities = append(cities, randomCity)
	}

	for _, city := range cities {
		if _, exists := cityFeatures[city.ID]; !exists {
			feature, err := getBestFittingGeoJSONFeature(city.Name, city.Country)
			if err != nil {
				return nil, err
			}
			cityFeatures[city.ID] = feature
		}
	}

	for i := range cities {
		point, err := RandomPosForCity(cityFeatures[cities[i].ID])
		if err != nil {
			fmt.Println(err)
			i--
		}
		positions = append(positions, point)
	}

	if len(positions) != count {
		return nil, errors.New(fmt.Sprintf("could not create %d rounds, created %d", count, len(positions)))
	}
	return
}

func queryWikiData(query string) *sparql.Results {
	fmt.Println("Wikidata query: " + query)
	repo, err := sparql.NewRepo(wikiDataUrl)
	if err != nil {
		log.Fatal(err)
	}
	results, err := repo.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Wikidata results: %v\n", results.Results.Bindings)
	return results
}
