package pkg

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"strconv"
	"strings"
)

const (
	continentsTable = "GeoContinents"
	citiesTable     = "GeoCities"
	countriesTable  = "GeoCountries"
)

type CountryName struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Countries struct {
	Countries []CountryName `json:"countries"`
}

type City struct {
	Name    string
	Pop     int
	Country string
}

func GetCountries(continent string) (*Countries, error) {
	returnCountries := &Countries{}

	params := &dynamodb.GetItemInput{
		TableName: aws.String(continentsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"continent": {
				S: aws.String(continent),
			},
		},
	}
	res, err := DynamoClient.GetItem(params)

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error getting countries for continent: %v", err))
	}

	for _, countryItem := range res.Item["countries"].L {
		country := countryItem.M
		if len(country) == 0 {
			continue
		}
		returnCountries.Countries = append(returnCountries.Countries, CountryName{
			Code: aws.StringValue(country["code"].S),
			Name: aws.StringValue(country["name"].S),
		})
	}
	return returnCountries, nil
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
	countryName, err := getCountryName(countryCode)
	if err != nil {
		return nil, err
	}
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

func getCountryName(countryCode string) (string, error) {
	params := &dynamodb.GetItemInput{
		TableName: aws.String(countriesTable),
		Key: map[string]*dynamodb.AttributeValue{
			"code": {
				S: aws.String(strings.ToUpper(countryCode)),
			},
		},
	}

	res, err := DynamoClient.GetItem(params)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error getting country: %v", err))
	}
	return aws.StringValue(res.Item["name"].S), nil
}
