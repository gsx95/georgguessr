package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/knakk/sparql"
	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"log"
	"strconv"
	"strings"
	"time"
	_ "embed"
)

var minPopulation = map[string]int{
	"pop_gt_100k": 100000,
	"pop_gt_500k": 500000,
	"pop_gt_1kk":  1000000,
	"pop_gt_5kk":  5000000,
}

//go:embed sparql/getCountryByCode.query
var getCountryByCode string
//go:embed sparql/getCitiesByCountryAndPop.query
var getCityByCountryAndPop string
//go:embed sparql/getRandomCityForPop.query
var getRandomCityForPop string

const wikiDataUrl = "https://query.wikidata.org/sparql"

type RoomWithPredefinedArea struct {
	pkg.Room
	Country   string `json:"country"`
	City      string `json:"city"`
}

type CreatorPredefinedCities struct {}


func (cr *CreatorPredefinedCities) CreateRoom(reqBody string) (string, error){
	room := RoomWithPredefinedArea{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	country := room.Country
	cities := room.City

	positions := positions{}

	randomPositions, err := cr.randomPositionByArea(country, cities, room.Rounds +additionalCreationTries)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("No point could be generated: %v", err))
	}

	for i, pos := range randomPositions  {
		positions.Pos = append(positions.Pos, newRoundPos(i, pos.Lat(), pos.Lon()))
	}

	streetViews, err := getStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room.Room, *streetViews)

	return createRoom(&room.Room)
}


// Returns error if no valid point could be generated
func (cr *CreatorPredefinedCities) randomPositionByArea(country string, cities string, count int) (positions []*orb.Point, err error) {

	minPop := minPopulation[cities]
	if country != "all" {
		countryData, err := cr.queryWikiData(fmt.Sprintf(getCountryByCode, strings.ToUpper(country)))
		if err != nil {
			return nil, err
		}
		countryEntity := countryData.Results.Bindings[0]["country"].Value
		entityParts := strings.Split(countryEntity, "/")
		countryID := entityParts[len(entityParts)-1]

		query := fmt.Sprintf(getCityByCountryAndPop, countryID, minPop, time.Now().String())

		results, err := cr.queryWikiData(query)
		if err != nil {
			return nil, err
		}
		var citiesSlice []*place
		bindings := results.Results.Bindings
		for _, b := range bindings {
			cityPop, err := strconv.Atoi(b["maxPopulation"].Value)
			cityName := b["cityLabel"].Value
			locationString := b["location"].Value
			if err != nil {
				log.Fatal(err)
			}
			if cityPop > minPop {
				pos, err := cr.wikiDataStringToPos(locationString)
				if err != nil {
					return nil, err
				}
				newCity := cr.newCity(cityName, country, cityPop, *pos)
				citiesSlice = append(citiesSlice, newCity)
			}
		}
		return cr.randomPosForCities(citiesSlice, count)
	}

	query := fmt.Sprintf(getRandomCityForPop, minPop, time.Now().String())
	results, err := cr.queryWikiData(query)
	if err != nil {
		return nil, err
	}
	res := results.Results.Bindings
	var possibleCities []*place
	for _, result := range res {
		cityName := result["cityLabel"].Value
		countryName := result["countryLabel"].Value
		locationString := result["location"].Value
		pop, err := strconv.Atoi(result["population"].Value)
		if err != nil {
			fmt.Printf("error while trying to convert wikiData pop %s to int\n", result["population"].Value)
			continue
		}
		pos, err := cr.wikiDataStringToPos(locationString)
		if err != nil {
			fmt.Println(err)
			continue
		}
		possibleCities = append(possibleCities, cr.newCity(cityName, countryName, pop, *pos))
	}
	return cr.randomPosForCities(possibleCities, count)
}

func (cr *CreatorPredefinedCities) randomPosForCities(possibleCities []*place, count int) (positions []*orb.Point, err error) {

	var cities []*place
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
		point, err := randomPosForCity(cityFeatures[cities[i].ID], cities[i].Pos)
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

func (cr *CreatorPredefinedCities) newCity(name, country string, pop int, pos position) *place {
	return &place{
		Pop: pop,
		Name: name,
		Country: country,
		ID: fmt.Sprintf("%s_%s_%d", name, country, pop),
		Pos: pos,
	}
}

func (cr *CreatorPredefinedCities) queryWikiData(query string) (*sparql.Results, error) {
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

func (cr *CreatorPredefinedCities) wikiDataStringToPos(queryResult string) (*position, error) { // e.g.  Point(7.099722222 50.733888888)
	coordinatesTxt := strings.Trim(queryResult, "Point()")
	coordinates := strings.Split(coordinatesTxt, " ")

	lng, err := strconv.ParseFloat(coordinates[0], 64)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Could not parse longitude %f from position %s", lng, coordinatesTxt))
	}
	lat, err := strconv.ParseFloat(coordinates[1], 64)
	if err != nil {
		return nil, pkg.InternalErr(fmt.Sprintf("Could not parse longitude %f from position %s", lng, coordinatesTxt))
	}

	return &position{
		Lng: lng,
		Lat: lat,
	}, nil

}