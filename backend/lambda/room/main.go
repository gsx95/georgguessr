package main

import (
	"encoding/json"
	"georgguessr.com/lambda-room/creation"
	"georgguessr.com/lambda-room/db"
	"georgguessr.com/pkg"
	"github.com/aws/aws-lambda-go/events"
	"log"
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
		"/panoramas": {
			POST: HandlePostPanoramas,
		},
	}

	pkg.StartLambda(methods)
}

func HandleRoomExists(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	defer pkg.LogDuration(pkg.Track())
	roomID := getRoomIdFromRequest(request)
	if db.RoomExists(roomID) {
		return pkg.GenerateResponse(ExistsResponse{true}, 200)
	}
	return pkg.GenerateResponse(ExistsResponse{false}, 404)
}

func HandleGetRoom(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	defer pkg.LogDuration(pkg.Track())
	roomID := getRoomIdFromRequest(request)
	room, err := pkg.GetRoom(roomID)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.GenerateResponse(room, 200)
}

func HandlePostRoom(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	defer pkg.LogDuration(pkg.Track())
	createType := request.QueryStringParameters["type"]
	log.Println(createType)
	creator, err := creation.NewCreator(createType)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	roomId, genPositions, err := creator.CreateRoom(request.Body)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.GenerateResponse(RoomCreationResponse{
		RoomId:       roomId,
		GenPositions: *genPositions,
	}, 200)
}

func HandlePostPanoramas(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	defer pkg.LogDuration(pkg.Track())
	var panoramasData db.PanoramaData
	err := json.Unmarshal([]byte(request.Body), &panoramasData)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	err = db.WritePanosToRoom(panoramasData.RoomId, panoramasData)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.OkResponse()
}

func getRoomIdFromRequest(request events.APIGatewayProxyRequest) string {
	defer pkg.LogDuration(pkg.Track())
	roomID := request.PathParameters["roomID"]
	if roomID == "" {
		pkg.PanicBadRequest("no room id given")
	}
	return roomID
}
