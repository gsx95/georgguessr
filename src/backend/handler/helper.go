package handler

import "github.com/aws/aws-lambda-go/events"

func GenerateResponse(body string, status int) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body: body,
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Access-Control-Allow-Credentials" : "true",
		},
	}
}
