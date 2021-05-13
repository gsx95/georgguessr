package handler

import (
	"backend/data"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

// TODO: why did I seperate 'Room' and 'Game', they are the same thing in our dynamodb
// split them in two different data structures entirely OR just merge them in the codebase
// currently, the methods here are just for creating or joining a game.
// Maybe create a separate 'GameLobby'? Or just merge the methods here into the game.go and rename references from 'Room' to 'Game'

func HandleRoomExists(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return GenerateResponse("no room id given", 400)
	}
	exists := data.RoomExists(roomID)
	if exists {
		return GenerateResponse("{\"exists\": true}", 200)
	}
	return GenerateResponse("{\"exists\": false}", 404)
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
	case "places": id, err = data.CreateRoomWithPlaces(request.Body)
	case "custom": id, err = data.CreateRoomWithCustomAreas(request.Body)
	default: err = errors.New("geo type was " + geoType)
	}

	if err != nil {
		return GenerateResponse(err.Error(), 400)
	}
	return GenerateResponse(id, 201)

}