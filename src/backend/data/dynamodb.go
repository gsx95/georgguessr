package data

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
	"strconv"
)

const roomsTable = "GeorgGuessrRooms"
const continentsTable = "GeoContinents"

var (
	DynamoClient dynamodbiface.DynamoDBAPI
)

type Room struct {
	ID         string `json:"id,omitempty"`
	Name       string `json:"name"`
	Players    int    `json:"players"`
	MaxPlayers int    `json:"maxPlayers"`
	Status     string `json:"status,omitempty"`
}

func GetRoom(roomID string) (*Room, error) {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
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
		fmt.Println(err)
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
			room.Players = 0
		} else {
			players, err := strconv.Atoi(roomItem["players"])
			if err != nil {
				return nil, err
			}
			room.Players = players
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

func CreateRoom(room Room) error {
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

	fmt.Println(res.Item)

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
