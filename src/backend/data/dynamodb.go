package data

import (
	"backend/helper"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"math/rand"
	"strconv"
	"strings"
)

const roomsTable = "GeorgGuessrRooms"
const continentsTable = "GeoContinents"
const countriesTable = "GeoCountries"
const citiesTable = "GeoCities"

var (
	DynamoClient dynamodbiface.DynamoDBAPI
)

type Room struct {
	ID            string        `json:"id,omitempty"`
	Name          string        `json:"name"`
	Players       []string           `json:"players"`
	MaxPlayers    int           `json:"maxPlayers"`
	Rounds        int           `json:"maxRounds"`
	Status        string        `json:"status,omitempty"`
	TimeLimit     int           `json:"timeLimit"`
	GeoBoundaries []GeoBoundary `json:"geoBoundaries"`
	GamesRounds   []GameRound   `json:"gameRounds"`

}

type GeoBoundary struct {
	GeoPoints []GeoPoint `json:"geoPoints"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type GameRound struct {
	No            int
	StartPosition GeoPoint          `json:"startPosition"`
	Scores        map[string]Guess `json:"scores"`
	PanoID        string            `json:"panoID"`
}

type Guess struct {
	Distance int  `json:"distance"`
	Position GeoPoint  `json:"guess"`
}

type City struct {
	Name    string
	Pop     int
	Country string
}

func PutGuess(gameID, username string, round int, guess Guess) error {

	guessMap, err := dynamodbattribute.Marshal(guess)
	if err != nil {
		return err
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(roomsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(gameID),
			},
		},
		UpdateExpression: aws.String(fmt.Sprintf("SET gameRounds[%d].scores.#username = :score", round)),
		ExpressionAttributeNames: map[string]*string{
			"#username": aws.String(username),
		},
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":score": guessMap,
		},
	}

	result, err := DynamoClient.UpdateItem(input)
	if err != nil {
		fmt.Println(result)
		return errors.New(fmt.Sprintf("%v  username: %s", err, username))
	}
	return nil
}

func PutUsername(gameID, username string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(roomsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(gameID),
			},
		},
		UpdateExpression:    aws.String("SET players = list_append(if_not_exists(players, :emptylist), :username)"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":emptylist": {
				L: []*dynamodb.AttributeValue{},
			},
			":username": {
				L: []*dynamodb.AttributeValue{
					{
						S: aws.String(username),
					},
				},
			},
		},
	}

	result, err := DynamoClient.UpdateItem(input)
	if err != nil {
		fmt.Println(result)
		return err
	}
	return nil
}

func PutPanoID(roomID string, round int, panoID string) error {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(roomsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(roomID),
			},
		},
		UpdateExpression:    aws.String(fmt.Sprintf("set gameRounds[%d].panoID = :item", round-1)),
		ConditionExpression: aws.String(fmt.Sprintf("attribute_not_exists(gameRounds[%d].panoID)", round-1)),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":item": {
				S: aws.String(panoID),
			},
		},
	}

	result, err := DynamoClient.UpdateItem(input)
	if err != nil {
		fmt.Println(result)
		return err
	}
	return nil
}

func GetRoom(roomID string) (*Room, error) {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(roomID),
			},
		},
		TableName: aws.String(roomsTable),
	}

	result, err := DynamoClient.GetItem(input)
	if err != nil {
		return nil, err
	}

	item := new(Room)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, err
	}
	return item, nil
}

func GetAvailableRooms() ([]*Room, error) {

	exprAttrValues := map[string]*dynamodb.AttributeValue{
		":creating": {
			S: aws.String("creating"),
		},
		":waiting": {
			S: aws.String("waiting"),
		},
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(roomsTable),
		FilterExpression:          aws.String("#status = :creating or #status = :waiting"),
		ExpressionAttributeValues: exprAttrValues,
		ExpressionAttributeNames: map[string]*string{
			"#status": aws.String("status"),
		},
	}

	result, err := DynamoClient.Scan(input)
	if err != nil {
		return nil, err
	}

	var items []map[string]string
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &items)
	if err != nil {
		return nil, err
	}

	var rooms []*Room
	for _, roomItem := range items {
		room := &Room{
			ID:     roomItem["id"],
			Status: roomItem["status"],
			Name:   roomItem["name"],
		}
		if roomItem["players"] == "" {
			room.Players = []string{}
		} else {
			players := roomItem["players"]
			//TODO:
			fmt.Println(players)
			room.Players = []string{}
		}
		maxPlayers, err := strconv.Atoi(roomItem["maxPlayers"])
		if err != nil {
			return nil, err
		}
		room.MaxPlayers = maxPlayers
		rooms = append(rooms, room)
	}
	return rooms, nil
}

func writeRoomToDB(room Room) error {
	av, err := dynamodbattribute.MarshalMap(room)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(roomsTable),
	}

	_, err = DynamoClient.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}

type CountryName struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type Countries struct {
	Countries []CountryName `json:"countries"`
}

func GetRandomCityName(minPop, maxPop int, countries map[string]bool) (string, string, error) {

	params := &dynamodb.ScanInput{
		TableName:            aws.String(citiesTable),
		ProjectionExpression: aws.String("biggest, country"),
	}
	res, err := DynamoClient.Scan(params)
	if err != nil {
		return "", "", err
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
	ranNum := helper.GetRandom(0, max)
	ranCountry := allList[ranNum]
	cities, err := getCities(ranCountry.Biggest, minPop, maxPop)
	if err != nil {
		return "", "", errors.New("something went wrong when getting cities " + err.Error())
	}
	city := cities[rand.Intn(len(cities))]
	return city.Name, city.Country, nil
}

func GetCountryName(countryCode string) (string, error) {
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
		return "", err
	}
	return aws.StringValue(res.Item["name"].S), nil

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
		return nil, err
	}

	countryCode := aws.StringValue(res.Item["country"].S)
	countryName, err := GetCountryName(countryCode)
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
		return nil, err
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
