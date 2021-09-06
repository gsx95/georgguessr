package main

import (
	"georgguessr.com/lambda-room/creation"
	"georgguessr.com/lambda-room/db"
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
	if db.RoomExists(roomID) {
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
	defer pkg.Duration(pkg.Track("HandlePostRoom [" + createType + "]"))

	creator, err := creation.NewCreator(createType)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	response, err := creator.CreateRoom(request.Body)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.StringResponse(response, 200)

}

func getRoomIdFromRequest(request events.APIGatewayProxyRequest) string {
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		pkg.PanicBadRequest("no room id given")
	}
	return roomID
}