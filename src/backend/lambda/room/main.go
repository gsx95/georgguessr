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
	var status int
	switch geoType {
	case "list":
		id, err, status = CreateRoomWithPredefinedArea(request.Body)
	case "unlimited":
		id, err, status = CreateRoomUnlimited(request.Body)
	case "places":
		id, err, status = CreateRoomWithPlaces(request.Body)
	case "custom":
		id, err, status = CreateRoomWithCustomAreas(request.Body)
	default:
		err = errors.New("geo type was " + geoType)
	}

	if err != nil {
		return pkg.GenerateResponse(err.Error(), status)
	}
	return pkg.GenerateResponse(id, status)

}
