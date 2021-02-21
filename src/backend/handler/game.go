package handler

import (
	"backend/data"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"strconv"
)

func HandlePostPanoIDs(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]
	round, err := strconv.Atoi(request.PathParameters["round"])
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}

	panoId := request.Body

	err = data.PutPanoID(gameID, round, panoId)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return GenerateResponse("ok", 200)

}

func HandleGetGamePosition(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]
	round, err := strconv.Atoi(request.PathParameters["round"])
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}
	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	rounds := game.GamesRounds
	if len(rounds) < round {
		return GenerateResponse("no more rounds", 400)
	}
	r := rounds[round - 1]

	bytes, err := json.Marshal(struct{
		PanoID string `json:"panoId"`
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	}{
		r.PanoID,
		r.StartPosition.Lat,
		r.StartPosition.Lon,
	})
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}

func HandleGetGameStats(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]

	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}
	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	bytes, err := json.Marshal(struct{
		Rounds int `json:"rounds"`
		Name string `json:"name"`
		MaxPlayers int `json:"maxPlayers"`
		Players int `json:"players"`
		Status string `json:"status"`
		TimeLimit int `json:"timeLimit"`
	}{
		game.Rounds,
		game.Name,
		game.MaxPlayers,
		game.Players,
		game.Status,
		game.TimeLimit,
	})
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}
