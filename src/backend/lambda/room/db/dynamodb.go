package db

import (
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func RoomExists(roomID string) bool {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(roomID),
			},
		},
		TableName: aws.String(pkg.RoomsTable),
	}

	result, err := pkg.DynamoClient.GetItem(input)
	if err != nil || result == nil || len(result.Item) == 0 {
		return false
	}
	return true
}

func WriteRoomToDB(room pkg.Room) error {

	av, err := pkg.Encoder.Encode(room)
	if err != nil {
		return pkg.InternalErr(fmt.Sprintf("Error marshalling room: %v", err))
	}

	input := &dynamodb.PutItemInput{
		Item:      av.M,
		TableName: aws.String(pkg.RoomsTable),
	}

	_, err = pkg.DynamoClient.PutItem(input)
	if err != nil {
		pkg.InternalErr(fmt.Sprintf("Error putting room: %v", err))
	}
	return nil
}