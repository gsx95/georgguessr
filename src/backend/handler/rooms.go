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

	type RoomInfo struct {
		ID            string        `json:"id,omitempty"`
		Name          string        `json:"name"`
		Players       int      		`json:"players"`
		MaxPlayers    int           `json:"maxPlayers"`
		Status        string        `json:"status,omitempty"`
	}

	roomInfos := make([]RoomInfo, 0)
	for _, room := range rooms {
		roomInfos = append(roomInfos, RoomInfo{
			ID: room.ID,
			Name: room.Name,
			Players: len(room.Players),
			MaxPlayers: room.MaxPlayers,
			Status: room.Status,
		})
	}

	byteRooms, err := json.Marshal(roomInfos)
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
	case "places": id, err = data.CreateRoomWithPlaces(request.Body)
	default: err = errors.New("geo type was " + geoType)
	}

	if err != nil {
		return GenerateResponse(err.Error(), 400)
	}
	return GenerateResponse(id, 201)

}