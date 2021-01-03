package main

import (
	"backend/pkg"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"os"
	"strings"
)

func handleGetRoom(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return events.APIGatewayProxyResponse{
			Body:       "no room id given.",
			StatusCode: 400,
		}, nil
	}
	room, err := data.GetRoom(roomID)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       "Room not found.",
			StatusCode: 404,
		}, nil
	}

	bytes, err := json.Marshal(room)
	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       "Room not found.",
			StatusCode: 404,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       string(bytes),
		StatusCode: 200,
	}, nil
}

func handlePostRoom(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	stringBody := request.Body
	var body map[string]string
	err := json.Unmarshal([]byte(stringBody), &body)

	name := body["name"]
	if name == "" {
		return events.APIGatewayProxyResponse{
			Body:       "no name provided in body",
			StatusCode: 400,
		}, nil
	}

	id := generateRoomID()

	err = data.CreateRoom(data.Room{
		ID: id,
		Name: name,
	})

	if err != nil {
		fmt.Println(err)
		return events.APIGatewayProxyResponse{
			Body:       "server error",
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       id,
		StatusCode: 201,
	}, nil
}

func handleGetAvailableRooms(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		Body:       "not implemented",
		StatusCode: 501,
	}, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func generateRoomID() string {
	b := make([]byte, 40)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

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

	return events.APIGatewayProxyResponse{
		Body:       "operation not supported",
		StatusCode: 400,
	}, nil
}

func main() {
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
