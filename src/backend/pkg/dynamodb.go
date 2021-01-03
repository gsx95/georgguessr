package data

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbiface"
)


const tableName = "GeorgGuessrRooms"

var (
	DynamoClient dynamodbiface.DynamoDBAPI
)

type Room struct {
	ID     string `json:"ID"`
	Name string `json:"name"`
}

func GetRoom(roomID string) (*Room, error) {

	input := &dynamodb.GetItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"ID": {
				S: aws.String(roomID),
			},
		},
		TableName: aws.String(tableName),
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

func CreateRoom(room Room)  error {
	av, err := dynamodbattribute.MarshalMap(room)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
	}

	_, err = DynamoClient.PutItem(input)
	if err != nil {
		return err
	}
	return nil
}