package main

import (
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

func HandleRoomExists(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	roomID := getRoomIdFromRequest(request)
	if RoomExists(roomID) {
		return pkg.GenerateResponse(ExistsResponse{true}, 200)
	}
	return pkg.GenerateResponse(ExistsResponse{false}, 404)
}

func HandleGetRoom(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	roomID := getRoomIdFromRequest(request)
	room, err := pkg.GetRoom(roomID)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.GenerateResponse(room, 200)
}

func HandlePostRoom(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {

	createType := request.QueryStringParameters["type"]
	switch createType {
	case "list":
		resp, err := CreateRoomWithPredefinedArea(request.Body)
		if err != nil {
			return pkg.ErrorResponse(err)
		}
		return pkg.StringResponse(resp, 200)
	case "unlimited":
		resp, err := CreateRoomUnlimited(request.Body)
		if err != nil {
			return pkg.ErrorResponse(err)
		}
		return pkg.StringResponse(resp, 200)
	case "places":
		resp, err := CreateRoomWithPlaces(request.Body)
		if err != nil {
			return pkg.ErrorResponse(err)
		}
		return pkg.StringResponse(resp, 200)
	case "custom":
		resp, err := CreateRoomWithCustomAreas(request.Body)
		if err != nil {
			return pkg.ErrorResponse(err)
		}
		return pkg.StringResponse(resp, 200)
	default:
		pkg.PanicBadRequest(fmt.Sprintf("Creation type %s not recognized", createType))
	}
	panic(fmt.Sprintf("Unknown error, creation type was %v", createType))
}

func getRoomIdFromRequest(request events.APIGatewayProxyRequest) string {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		pkg.PanicBadRequest("no room id given")
	}
	return roomID
}