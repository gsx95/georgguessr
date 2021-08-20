package pkg

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"os"
	"time"
)

type HandlerFunc func(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse

type MethodHandler struct {
	GET  HandlerFunc
	POST HandlerFunc
}

type MethodHandlers map[string]*MethodHandler

var (
	methodHandlers MethodHandlers
)

func StartLambda(methods MethodHandlers) {

	methodHandlers = methods

	rand.Seed(time.Now().UnixNano())
	region := os.Getenv("AWS_REGION")
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(region)},
	)
	if err != nil {
		return
	}

	_, isLocal := os.LookupEnv("AWS_SAM_LOCAL")

	config := aws.NewConfig().WithRegion(region)

	if isLocal {
		config.Endpoint = aws.String("http://dynamodb:8000")
	}
	DynamoClient = dynamodb.New(awsSession, config)

	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	handler, validPath := methodHandlers[request.Resource]

	if !validPath {
		return GenerateResponse("operation not supported", 400), nil
	}

	if request.HTTPMethod == "GET" {
		if handler.GET == nil {
			return GenerateResponse("GET method not supported", 400), nil
		}
		return handler.GET(request), nil
	}

	if request.HTTPMethod == "POST" {
		if handler.POST == nil {
			return GenerateResponse("POST method not supported", 400), nil
		}
		return handler.POST(request), nil
	}

	return GenerateResponse("method not supported", 400), nil
}

func GenerateResponse(body string, status int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}
}
