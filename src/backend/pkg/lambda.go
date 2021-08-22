package pkg

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"math/rand"
	"os"
	"time"
)

type HandlerFunc func(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse

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
		panic(fmt.Sprintf("Error creating aws session: %v", err))
	}

	_, isLocal := os.LookupEnv("AWS_SAM_LOCAL")

	config := aws.NewConfig().WithRegion(region)

	if isLocal {
		config.Endpoint = aws.String("http://dynamodb:8000")
	}
	DynamoClient = dynamodb.New(awsSession, config)

	lambda.Start(handler)
}

func handler(request events.APIGatewayProxyRequest) (response events.APIGatewayProxyResponse, execError error) {
	handler, _ := methodHandlers[request.Resource]

	defer func() {
		if r := recover(); r != nil {
			switch err := r.(type) {
			case error:
				response = *internalErrorResponse(err)
			case string:
				response = *internalErrorResponse(errors.New(err))
			default:
				response = *internalErrorResponse(errors.New(fmt.Sprintf("%v", err)))
			}
		}
	}()

	if request.HTTPMethod == "GET" {
		response = *handler.GET(request)
	}else if request.HTTPMethod == "POST" {
		response = *handler.POST(request)
	} else {
		response = *notFoundResponse("method not supported.")
	}
	return
}
