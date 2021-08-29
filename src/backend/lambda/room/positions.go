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
}

func (p *Place) getID() string {
	return fmt.Sprintf("%s_%s_%d", p.Name, p.Country, p.Pop)
}

func NewCity(name, country string, pop int) *Place {
	return &Place{
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
			if err != nil {
				log.Fatal(err)
			}
			if cityPop > minPop {
				citiesSlice = append(citiesSlice, NewCity(b["cityLabel"].Value, country, cityPop))
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
		pop, err := strconv.Atoi(result["population"].Value)
		if err != nil {
			continue
		}
		possibleCities = append(possibleCities, NewCity(cityName, countryName, pop))
	}
	return randomPosForCities(possibleCities, count)
}

// Returns error if no valid point could be generated
func RandomPosForCity(feature *geojson.Feature) (point *orb.Point, err error) {

	pointValid := false
	box := feature.BBox.Bound()

	for !pointValid {
		lat := pkg.GetRandomFloat(box.Min.Lat(), box.Max.Lat())
		lon := pkg.GetRandomFloat(box.Min.Lon(), box.Max.Lon())
		point = &orb.Point{lon, lat}
		pointValid, err = isPointInsidePolygon(feature, point)
		if err != nil {
			fmt.Println(err)
		}
	}

	return point, nil
}

func getBestFittingGeoJSONFeature(city, country string) (*geojson.Feature, error) {
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

	var placeFeatures []*geojson.Feature

	for _, feature := range featureCollection.Features {
		featureCategory := strings.ToLower(feature.Properties.MustString("category"))
		if featureCategory == "place" || featureCategory == "boundary" {
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

func isPointInsidePolygon(feature *geojson.Feature, point *orb.Point) (bool, error) {

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
		p := orb.Point{}
		p[0] = point.Lon
		p[1] = point.Lat
		ring = append(ring, p)
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
			feature, err := getBestFittingGeoJSONFeature(city.Name, city.Country)
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
		point, err := RandomPosForCity(cityFeatures[cities[i].ID])
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
