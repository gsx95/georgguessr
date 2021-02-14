package handler

import (
	"backend/data"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

func HandleGetAvailableRooms(_ events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	rooms, err := data.GetAvailableRooms()
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}

	byteRooms, err := json.Marshal(rooms)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(byteRooms), 200)
}


func HandleGetRoom(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return GenerateResponse("no room id given", 400)
	}
	room, err := data.GetRoom(roomID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	bytes, err := json.Marshal(room)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}

func HandlePostRoom(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {

	geoType := request.QueryStringParameters["type"]
	var id string
	var err error
	switch geoType {
	case "list": id, err = data.CreateRoomWithPredefinedArea(request.Body)
	case "unlimited": id, err = data.CreateRoomUnlimited(request.Body)
	default: err = errors.New("geo type was " + geoType)
	}

	if err != nil {
		return GenerateResponse(err.Error(), 400)
	}
	return GenerateResponse(id, 201)

}