package main

import (
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/knakk/sparql"
	"github.com/mmcloughlin/spherand"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
	_ "embed"
)

const getCityBoundariesUrl = "https://nominatim.openstreetmap.org/search.php?q=%s+%s&polygon_geojson=1&format=geojson"
const wikiDataUrl = "https://query.wikidata.org/sparql"

var minPopulation = map[string]int{
	"pop_gt_100k": 100000,
	"pop_gt_500k": 500000,
	"pop_gt_1kk": 1000000,
	"pop_gt_5kk": 5000000,
}

type CountryName struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type City struct {
	Name    string
	Pop     int
	Country string
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

func RandomPosForCity(city *City) (point *orb.Point, err error) {
	feature, err := getAdminFeatureForCity(city.Name, city.Country)
	if err != nil {
		return nil, err
	}

	bbound := feature.BBox.Bound()

	pointValid := false

	for !pointValid {
		lat := pkg.GetRandomFloat(bbound.Min.Lat(), bbound.Max.Lat())
		lon := pkg.GetRandomFloat(bbound.Min.Lon(), bbound.Max.Lon())
		point = &orb.Point{lon, lat}
		pointValid = isPointInsidePolygon(feature, point)
	}

	return point, nil
}

// TODO: this fails for places that are not "administrative" features, e.g. the village 'Afrikaskop'
// include all valid feature types of openstreetmaps, see branch "single-point-places"
func getAdminFeatureForCity(city, country string) (*geojson.Feature, error) {
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

	var adminFeature *geojson.Feature
	adminImportance := 0.0

	for _, feature := range featureCollection.Features {
		featureType := strings.ToLower(feature.Properties.MustString("type"))
		if featureType == "administrative" {
			importance := feature.Properties.MustFloat64("importance", 0)
			if importance >= adminImportance {
				adminFeature = feature
				adminImportance = importance
			}
		}
	}

	return adminFeature, nil
}

func isPointInsidePolygon(feature *geojson.Feature, point *orb.Point) bool {

	var polygon orb.Polygon

	// if its a polygon, we only care for the polygon with more boundary nodes
	multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon)
	if isMulti {
		polyLen := 0
		for _, pol := range multiPoly {
			if len(pol[0]) >= polyLen {
				polyLen = len(pol[0])
				polygon = pol
			}
		}
	} else {
		polygon, _ = feature.Geometry.(orb.Polygon)
	}

	if planar.PolygonContains(polygon, *point) {
		return true
	}
	return false
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
		countryMaxCityPop, err := strconv.Atoi(b["maxPopulation"].Value)
		if err != nil {
			log.Fatal(err)
		}
		if countryMaxCityPop > minPop {
			cities = append(cities, &City{
				Name: b["cityLabel"].Value,
				Country: country,
			})
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
		possibleCities = append(possibleCities, &City{
			Name: cityName,
			Country: countryName,
		})
	}
	return randomPosForCities(possibleCities, count)
}

func randomPosForCities(possibleCities []*City, count int) (positions []*orb.Point, err error){
	for i := 0;i < count; i++ {
		randomCity := possibleCities[pkg.GetRandom(0, len(possibleCities) - 1)]
		point, err := RandomPosForCity(randomCity)
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
