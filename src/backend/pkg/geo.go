package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"strconv"
	"strings"
)

import _ "embed"

// countries from https://github.com/annexare/Countries

//go:embed "countries.json"
var countriesData string

const (
	citiesTable = "GeoCities"
)

type CountryName struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type City struct {
	Name    string
	Pop     int
	Country string
}

type CountryData struct {
	Name      string   `json:"name"`
	Native    string   `json:"native"`
	Phone     string   `json:"phone"`
	Continent string   `json:"continent"`
	Capital   string   `json:"capital"`
	Currency  string   `json:"currency"`
	Languages []string `json:"languages"`
}

func GetCountries(continent string) (countries []*CountryName) {
	data := map[string]CountryData{}
	json.Unmarshal([]byte(countriesData), &data)
	for countryCode, countryData := range data {
		if countryData.Continent == continent {
			countries = append(countries, &CountryName{
				Code: countryCode,
				Name: countryData.Name,
			})
		}
	}
	return
}

func GetRandomCityName(minPop, maxPop int, countries map[string]bool) (string, string, error) {

	params := &dynamodb.ScanInput{
		TableName:            aws.String(citiesTable),
		ProjectionExpression: aws.String("biggest, country"),
	}
	res, err := DynamoClient.Scan(params)
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("Error scanning cities table: %v", err))
	}

	type proj struct {
		Country string
		Biggest int
	}

	var allList []proj

	for _, item := range res.Items {
		biggest, _ := strconv.Atoi(aws.StringValue(item["biggest"].N))
		country := strings.ToLower(aws.StringValue(item["country"].S))
		if biggest < minPop {
			continue
		}
		if len(countries) != 0 {
			if _, isAllowed := countries[country]; !isAllowed {
				continue
			}
		}
		allList = append(allList, proj{
			Country: country,
			Biggest: biggest,
		})
	}

	max := len(allList) - 1
	ranNum := GetRandom(0, max)
	ranCountry := allList[ranNum]
	cities, err := getCities(ranCountry.Biggest, minPop, maxPop)
	if err != nil {
		return "", "", errors.New("Error getting cities " + err.Error())
	}
	city := cities[rand.Intn(len(cities))]
	return city.Name, city.Country, nil
}

func getCities(biggest, min, max int) ([]City, error) {
	params := &dynamodb.GetItemInput{
		TableName: aws.String(citiesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"biggest": {
				N: aws.String(strconv.Itoa(biggest)),
			},
		},
	}
	res, err := DynamoClient.GetItem(params)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error getting cities: %v", err))
	}

	countryCode := aws.StringValue(res.Item["country"].S)
	countryName := getCountryName(countryCode)
	citiesAttr := res.Item["cities"].L

	var cities []City

	for _, cityAttr := range citiesAttr {
		cityMap := cityAttr.M

		name := aws.StringValue(cityMap["name"].S)
		pop, _ := strconv.Atoi(aws.StringValue(cityMap["pop"].N))
		if min != -1 && pop < min {
			continue
		}
		if max != -1 && pop > max {
			continue
		}
		cities = append(cities, City{
			Name:    name,
			Pop:     pop,
			Country: countryName,
		})
	}
	return cities, nil
}

func getCountryName(countryCode string) string {
	countries := map[string]CountryData{}
	json.Unmarshal([]byte(countriesData), &countries)
	return countries[countryCode].Name
}
