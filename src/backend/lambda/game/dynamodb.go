package main

import (
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func PutGuess(gameID, username string, round int, guess pkg.Guess) error {

	guessMap, err := pkg.Encoder.Encode(guess)
	if err != nil {
		return errors.New(fmt.Sprintf("Error marshalling guess: %v", err))
	}

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(pkg.RoomsTable),
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

	_, err = pkg.DynamoClient.UpdateItem(input)
	if err != nil {
		return errors.New(fmt.Sprintf("error putting guess: %v username: %s", err, username))
	}
	return nil
}

func PutUsername(gameID, username string) error {

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(pkg.RoomsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(gameID),
			},
		},
		UpdateExpression: aws.String("SET players = list_append(if_not_exists(players, :emptylist), :username)"),
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

	_, err := pkg.DynamoClient.UpdateItem(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Error putting username: %v", err))
	}
	return nil
}

func PutPanoID(roomID string, round int, panoID string) error {

	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(pkg.RoomsTable),
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

	_, err := pkg.DynamoClient.UpdateItem(input)
	if err != nil {
		return errors.New(fmt.Sprintf("Error putting panorama ID: %v", err))
	}
	return nil
}
