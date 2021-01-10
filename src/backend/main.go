package main

import (
	"backend/pkg"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"os"
	"strings"
	"time"
)

func handleGetRoom(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return generateResponse("no room id given", 400), nil
	}
	room, err := data.GetRoom(roomID)
	if err != nil {
		return generateResponse(fmt.Sprintf("%v", err), 404), nil
	}
	bytes, err := json.Marshal(room)
	if err != nil {
		return generateResponse(fmt.Sprintf("%v", err), 500), nil
	}
	return generateResponse(string(bytes), 200), nil
}

func handlePostRoom(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	id, err := createRoom(request.Body)
	if err != nil {
		return generateResponse(fmt.Sprintf("%v", err), 500), nil
	}
	return generateResponse(id, 201), nil
}

func handleGetAvailableRooms(_ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	rooms, err := data.GetAvailableRooms()
	if err != nil {
		return generateResponse(fmt.Sprintf("%v", err), 500), nil
	}

	byteRooms, err := json.Marshal(rooms)
	if err != nil {
		return generateResponse(fmt.Sprintf("%v", err), 500), nil
	}
	return generateResponse(string(byteRooms), 200), nil
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if request.HTTPMethod == "OPTIONS" {
		return generateResponse("hi", 200), nil
	}

	fmt.Println(request.Path)

	if strings.HasPrefix(request.Path, "/rooms") {
		if request.HTTPMethod == "GET" {
			return handleGetRoom(request)
		}
		if request.HTTPMethod == "POST" {
			return handlePostRoom(request)
		}
	}

	if strings.HasPrefix(request.Path, "/available-rooms") {
		return handleGetAvailableRooms(request)
	}

	return generateResponse("operation not supported", 400), nil
}

func main() {
	rand.Seed(time.Now().UnixNano())
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return
	}
	data.DynamoClient = dynamodb.New(awsSession)
	lambda.Start(handler)
}

func createRoom(body string) (roomId string, err error) {

	room := data.Room{}
	err = json.Unmarshal([]byte(body), &room)
	if err != nil {
		return "", err
	}

	if room.Name == "" {
		return "", errors.New("no room name provided")
	}

	if room.MaxPlayers == 0 {
		return "", errors.New("zero players not possible")
	}

	id := generateRoomID()
	room.ID = id
	room.Players = 0
	room.Status = "creating"
	err = data.CreateRoom(room)
	if err != nil {
		return "", err
	}
	return id, nil
}

func generateResponse(body string, status int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body: body,
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Credentials" : "true",
		},
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRoomID() string {
	b := make([]byte, 40)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}