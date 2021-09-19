package pkg

import (
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"log"
)

func notFoundResponse(errorMsg string) *events.APIGatewayProxyResponse {
	return GenerateResponse(stringToJsonStruct(errorMsg), 404)
}

func badRequestResponse(errorMsg string) *events.APIGatewayProxyResponse {
	return GenerateResponse(stringToJsonStruct(errorMsg), 400)
}

func internalErrorResponse(err error) *events.APIGatewayProxyResponse {
	log.Println("Error: ")
	log.Println(err)
	return GenerateResponse(stringToJsonStruct("Internal Server Error"), 500)
}

func ErrorResponse(err error) *events.APIGatewayProxyResponse {
	if ErrType(err) == BAD_REQUEST {
		return badRequestResponse(err.Error())
	} else if ErrType(err) == NOT_FOUND {
		return notFoundResponse(err.Error())
	} else {
		return internalErrorResponse(err)
	}
}

func OkResponse() *events.APIGatewayProxyResponse {
	return GenerateResponse(stringToJsonStruct("ok"), 200)
}

func GenerateResponse(jsonData interface{}, status int) *events.APIGatewayProxyResponse {

	msg, err := json.Marshal(jsonData)
	if err != nil {
		return internalErrorResponse(err)
	}
	return &events.APIGatewayProxyResponse{
		Body:       string(msg),
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Origin":      "*",
			"Access-Control-Allow-Credentials": "true",
		},
	}
}

func stringToJsonStruct(msg string) interface{} {
	return struct {
		Msg string `json:"msg"`
	}{
		Msg: msg,
	}
}
