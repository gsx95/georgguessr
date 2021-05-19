package pkg

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)

var (
	DynamoClient dynamodbiface.DynamoDBAPI

	Encoder = dynamodbattribute.NewEncoder(func(e *dynamodbattribute.Encoder) {
		e.EnableEmptyCollections = true
	})
)

const RoomsTable = "GeorgGuessrRooms"

type Room struct {
	ID           string       `json:"id,omitempty"`
	Players      []string     `json:"players"`
	MaxPlayers   int          `json:"maxPlayers"`
	Rounds       int          `json:"maxRounds"`
	Status       string       `json:"status,omitempty"`
	TimeLimit    int          `json:"timeLimit"`
	Areas        [][]GeoPoint `json:"areas,omitempty"`
	GamesRounds  []GameRound  `json:"gameRounds"`
	PlayersCount int          `json:"playersCount,omitempty"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lng"`
}

type GameRound struct {
	No            int
	StartPosition GeoPoint         `json:"startPosition"`
	Scores        map[string]Guess `json:"scores"`
	PanoID        string           `json:"panoID,omitempty"`
}

type Guess struct {
	Distance int      `json:"distance"`
	Position GeoPoint `json:"guess"`
}

func GetRoom(roomID string) (*Room, error) {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(roomID),
			},
		},
		TableName: aws.String(RoomsTable),
	}

	result, err := DynamoClient.GetItem(input)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error getting room item: %v", err))
	}

	item := new(Room)
	err = dynamodbattribute.UnmarshalMap(result.Item, item)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Error unmarhsalling room item: %v", err))
	}
	item.PlayersCount = len(item.Players)
	return item, nil
}
