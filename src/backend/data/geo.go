package data

import (
	"errors"
	"fmt"
	"github.com/mmcloughlin/spherand"
)

func RandomPosition() (lat, lon float64) {
	lat, lon = spherand.Geographical()
	return
}

func RandomPositionByArea(continent string, country string, cities string) (lat, lon float64, err error) {

	countries := map[string]bool{}

	if country != "all" {
		countries[country] = true
	} else if continent != "all" {
		continentCountries, err := GetCountries(continent)
		if err != nil {
			return 0, 0, err
		}
		for _, c := range continentCountries.Countries {
			countries[c.Code] = true
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

func getRandomPosByArea(minPop, maxPop int, countries map[string]bool) (lat, lon float64, err error) {
	randomCity, err := GetRandomCityName(minPop, maxPop, countries)
	if err != nil {
		return 0, 0, err
	}
	fmt.Println(randomCity)
	// get boundary for city
	// return random point inside boundary
	return 135, 135, nil
}
