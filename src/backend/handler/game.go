package handler

import (
	"backend/data"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"strconv"
	"strings"
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
	r := rounds[round-1]

	bytes, err := json.Marshal(struct {
		Areas  [][]data.GeoPoint `json:"areas,omitempty"`
		PanoID string            `json:"panoId"`
		Lat    float64           `json:"lat"`
		Lon    float64           `json:"lon"`
	}{
		game.Areas,
		r.PanoID,
		r.StartPosition.Lat,
		r.StartPosition.Lon,
	})
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}

func HandleGetGuess(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]
	round, err := strconv.Atoi(request.PathParameters["round"])

	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}

	game, err := data.GetRoom(gameID)
	scores := game.GamesRounds[round-1].Scores

	bytes, err := json.Marshal(scores)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return GenerateResponse(string(bytes), 200)

}

func HandlePostGuess(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]
	username := request.PathParameters["username"]
	round, err := strconv.Atoi(request.PathParameters["round"])

	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}

	var guess data.Guess

	if err := json.Unmarshal([]byte(request.Body), &guess); err != nil {
		return GenerateResponse("Invalid guess body: "+err.Error(), 400)
	}

	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	scores := game.GamesRounds[round-1].Scores
	if _, alreadyExists := scores[username]; alreadyExists {
		return GenerateResponse("Already posted score for this round", 400)
	}

	err = data.PutGuess(gameID, username, round-1, guess)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 400)
	}

	return GenerateResponse("ok", 200)
}

func HandlePostPlayers(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]
	username := request.PathParameters["username"]

	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}
	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	for _, user := range game.Players {
		if strings.ToLower(user) == strings.ToLower(username) {
			return GenerateResponse("Username already in use", 400)
		}
	}

	err = data.PutUsername(gameID, username)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return GenerateResponse("ok", 200)
}

func HandleGetPlayers(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]

	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}
	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	bytes, err := json.Marshal(struct {
		Players []string `json:"players"`
	}{
		game.Players,
	})
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}

func HandleGetGameEndResults(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	gameID := request.PathParameters["gameID"]

	if gameID == "" {
		return GenerateResponse("no game id given", 400)
	}
	game, err := data.GetRoom(gameID)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	bytes, err := json.Marshal(struct {
		Rounds      int               `json:"rounds"`
		Players     []string          `json:"players"`
		Areas       [][]data.GeoPoint `json:"areas,omitempty"`
		GamesRounds []data.GameRound  `json:"gameRounds"`
	}{
		game.Rounds,
		game.Players,
		game.Areas,
		game.GamesRounds,
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
	bytes, err := json.Marshal(struct {
		Rounds     int    `json:"rounds"`
		MaxPlayers int    `json:"maxPlayers"`
		Players    int    `json:"players"`
		Status     string `json:"status"`
		TimeLimit  int    `json:"timeLimit"`
	}{
		game.Rounds,
		game.MaxPlayers,
		len(game.Players),
		game.Status,
		game.TimeLimit,
	})
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}
