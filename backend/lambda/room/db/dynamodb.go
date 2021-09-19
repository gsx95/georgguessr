package db

import (
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"time"
)

type PanoramaData struct {
	RoomId      string   `json:"roomId"`
	Panoramas []struct{
		Id string `json:"id"`
		Pos pkg.GeoPoint `json:"pos"`
	} `json:"panoramas"`
}

func WritePanosToRoom(roomID string, panosData PanoramaData) error {
	var rounds []pkg.GameRound
	for i, panoramaData := range panosData.Panoramas {
		gr := pkg.GameRound{
			No: i,
			StartPosition: pkg.GeoPoint{
				Lat: panoramaData.Pos.Lat,
				Lng: panoramaData.Pos.Lng,
			},
			PanoID: panoramaData.Id,
			Scores: map[string]pkg.Guess{},
		}
		rounds = append(rounds, gr)
	}

	gameRoundsEncoded, err := pkg.Encoder.Encode(rounds)
	if err != nil {
		return pkg.InternalErr(fmt.Sprintf("Error encoding game rounds: %v - %v", err, rounds))
	}

	updateInput := &dynamodb.UpdateItemInput{
		TableName: aws.String(pkg.RoomsTable),
		Key: map[string]*dynamodb.AttributeValue{
			"id": {
				S: aws.String(roomID),
			},
		},
		UpdateExpression: aws.String("SET gameRounds = :gRounds"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":gRounds": gameRoundsEncoded,
		},
		ConditionExpression: aws.String("attribute_not_exists(gameRounds)"),
	}
	_, err = pkg.DynamoClient.UpdateItem(updateInput)
	if err != nil {
		return pkg.InternalErr(fmt.Sprintf("Error updating item: %v", err))
	}
	return nil
}

func RoomExists(roomID string) bool {
	defer pkg.LogDuration(pkg.Track())
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
	room.TTL = time.Now().Unix() + (60 * 60 * 2)
	defer pkg.LogDuration(pkg.Track())
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