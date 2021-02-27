package main

import (
	"backend/data"
	h "backend/handler"
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



func webHandler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	if request.HTTPMethod == "OPTIONS" {
		return h.GenerateResponse("hi", 200), nil
	}

	if strings.HasPrefix(request.Path, "/rooms") {
		if request.HTTPMethod == "GET" {
			return h.HandleGetRoom(request), nil
		}
		if request.HTTPMethod == "POST" {
			return h.HandlePostRoom(request), nil
		}
	}

	if strings.HasPrefix(request.Path, "/game/pos") {
		return h.HandleGetGamePosition(request), nil
	}

	if strings.HasPrefix(request.Path, "/game/pano") {
		return h.HandlePostPanoIDs(request), nil
	}

	if strings.HasPrefix(request.Path, "/game/stats") {
		return h.HandleGetGameStats(request), nil
	}

	if strings.HasPrefix(request.Path, "/game/players") {
		if request.HTTPMethod == "GET" {
			return h.HandleGetPlayers(request), nil
		}
		return h.HandlePostPlayers(request), nil
	}

	if strings.HasPrefix(request.Path, "/available-rooms") {
		return h.HandleGetAvailableRooms(request), nil
	}

	if strings.HasPrefix(request.Path, "/countries") {
		return h.HandleGetCountriesInContinent(request), nil
	}

	return h.GenerateResponse("operation not supported", 400), nil
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
	lambda.Start(webHandler)
}
