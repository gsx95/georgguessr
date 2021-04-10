package data

import (
	"backend/helper"
	"errors"
	"fmt"
	"github.com/mmcloughlin/spherand"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
	"io/ioutil"
	"net/http"
	"strings"
)

const getCityBoundariesUrl = "https://nominatim.openstreetmap.org/search.php?q=%s+%s&polygon_geojson=1&format=geojson"

func RandomPosition() (lat, lon float64) {
	lat, lon = spherand.Geographical()
	return
}

func RandomPositionByArea(continent string, country string, cities string) (lat, lon float64, err error) {

	countries := map[string]bool{}

	if country != "all" {
		countries[strings.ToLower(country)] = true
	} else if continent != "all" {
		continentCountries, err := GetCountries(continent)
		if err != nil {
			return 0, 0, err
		}
		for _, c := range continentCountries.Countries {
			countries[strings.ToLower(c.Code)] = true
		}
	}

	switch cities {
	case "all":
		return getRandomPosByArea(0, -1, countries)
	case "pop_gt_100k":
		return getRandomPosByArea(100000, -1, countries)
	case "pop_gt_500k":
		return getRandomPosByArea(500000, -1, countries)
	case "pop_gt_1kk":
		return getRandomPosByArea(1000000, -1, countries)
	case "pop_gt_5kk":
		return getRandomPosByArea(5000000, -1, countries)
	case "pop_lt_100k":
		return getRandomPosByArea(0, 100000, countries)
	case "pop_lt_500k":
		return getRandomPosByArea(0, 500000, countries)
	case "pop_lt_1kk":
		return getRandomPosByArea(0, 1000000, countries)
	}

	return 0, 0, errors.New("unknown city type " + cities)
}

func RandomPosForCity(city, country string) (lat, lon float64, err error) {
	feature, err := getAdminFeatureForCity(city, country)
	if err != nil {
		return 0, 0, err
	}

	bbound := feature.BBox.Bound()

	pointValid := false
	var point orb.Point

	for !pointValid {
		lat := helper.GetRandomFloat(bbound.Min.Lat(), bbound.Max.Lat())
		lon := helper.GetRandomFloat(bbound.Min.Lon(), bbound.Max.Lon())
		point = orb.Point{lon, lat}
		pointValid = isPointInsidePolygon(feature, point)
	}

	return point.Lat(), point.Lon(), nil
}

func getRandomPosByArea(minPop, maxPop int, countries map[string]bool) (lat, lon float64, err error) {
	randomCity, country, err := GetRandomCityName(minPop, maxPop, countries)
	if err != nil {
		return 0, 0, err
	}
	return RandomPosForCity(randomCity, country)
}

func getAdminFeatureForCity(city, country string) (*geojson.Feature, error) {
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

func isPointInsidePolygon(feature *geojson.Feature, point orb.Point) bool {

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

	if planar.PolygonContains(polygon, point) {
		return true
	}
	return false
}
