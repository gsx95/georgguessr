package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-lambda-go/events"
)

func main() {

	methods := pkg.MethodHandlers{
		"/exists/{roomID}": {
			GET: HandleRoomExists,
		},
		"/rooms/{roomID}": {
			GET: HandleGetRoom,
		},
		"/rooms": {
			POST: HandlePostRoom,
		},
	}

	pkg.StartLambda(methods)
}

// TODO: why did I separate 'Room' and 'Game', they are the same thing in our dynamodb
// split them in two different data structures entirely OR just merge them in the codebase
// currently, the methods here are just for creating or joining a game.
// Maybe create a separate 'GameLobby'? Or just merge the methods here into the game package and rename references from 'Room' to 'Game'

func HandleRoomExists(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return pkg.GenerateResponse("no room id given", 400)
	}
	exists := RoomExists(roomID)
	if exists {
		return pkg.GenerateResponse("{\"exists\": true}", 200)
	}
	return pkg.GenerateResponse("{\"exists\": false}", 404)
}

func HandleGetRoom(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		return pkg.GenerateResponse("no room id given", 400)
	}
	room, err := pkg.GetRoom(roomID)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	bytes, err := json.Marshal(room)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

func HandlePostRoom(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {

	geoType := request.QueryStringParameters["type"]
	var id string
	var err error
	switch geoType {
	case "list":
		id, err = CreateRoomWithPredefinedArea(request.Body)
	case "unlimited":
		id, err = CreateRoomUnlimited(request.Body)
	case "places":
		id, err = CreateRoomWithPlaces(request.Body)
	case "custom":
		id, err = CreateRoomWithCustomAreas(request.Body)
	default:
		err = errors.New("geo type was " + geoType)
	}

	if err != nil {
		return pkg.GenerateResponse(err.Error(), 400)
	}
	return pkg.GenerateResponse(id, 201)

}
