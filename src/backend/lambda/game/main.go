package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-lambda-go/events"
	"strconv"
	"strings"
)

func main() {

	methods := pkg.MethodHandlers{
		"/game/pos/{gameID}/{round}": {
			GET: HandleGetGamePosition,
		},
		"/game/stats/{gameID}": {
			GET: HandleGetGameStats,
		},
		"/game/players/{gameID}": {
			GET: HandleGetPlayers,
		},
		"/game/players/{gameID}/{username}": {
			POST: HandlePostPlayers,
		},
		"/game/pano/{gameID}/{round}": {
			POST: HandlePostPanoIDs,
		},
		"/game/guess/{gameID}/{round}/{username}": {
			POST: HandlePostGuess,
		},
		"/game/guesses/{gameID}/{round}": {
			GET: HandleGetGuess,
		},
		"/game/endresults/{gameID}": {
			GET: HandleGetGameEndResults,
		},
	}

	pkg.StartLambda(methods)
}

func HandlePostPanoIDs(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	round, err := strconv.Atoi(request.PathParameters["round"])
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}

	panoId := request.Body

	err = PutPanoID(game.ID, round, panoId)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return pkg.GenerateResponse("ok", 200)
}

func HandleGetGuess(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	round, err := strconv.Atoi(request.PathParameters["round"])
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}

	scores := game.GamesRounds[round-1].Scores

	bytes, err := json.Marshal(scores)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

func HandlePostGuess(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	username := request.PathParameters["username"]
	round, err := strconv.Atoi(request.PathParameters["round"])
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}

	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	var guess pkg.Guess

	if err := json.Unmarshal([]byte(request.Body), &guess); err != nil {
		return pkg.GenerateResponse("Invalid guess body: "+err.Error(), 400)
	}
	scores := game.GamesRounds[round-1].Scores
	if _, alreadyExists := scores[username]; alreadyExists {
		return pkg.GenerateResponse("Already posted score for this round", 400)
	}

	err = PutGuess(game.ID, username, round-1, guess)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 400)
	}

	return pkg.GenerateResponse("ok", 200)
}

func HandlePostPlayers(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	username := request.PathParameters["username"]
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}

	for _, user := range game.Players {
		if strings.ToLower(user) == strings.ToLower(username) {
			return pkg.GenerateResponse("Username already in use", 400)
		}
	}

	err := PutUsername(game.ID, username)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	return pkg.GenerateResponse("ok", 200)
}

func HandleGetPlayers(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}
	bytes, err := json.Marshal(struct {
		Players []string `json:"players"`
	}{
		game.Players,
	})
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

func HandleGetGameEndResults(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}
	bytes, err := json.Marshal(pkg.Room{
		Rounds:      game.Rounds,
		Players:     game.Players,
		Areas:       game.Areas,
		GamesRounds: game.GamesRounds,
	})
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

type PositionResponse struct {
	Areas  [][]pkg.GeoPoint `json:"areas,omitempty"`
	PanoID string           `json:"panoId"`
	Lat    float64          `json:"lat"`
	Lon    float64          `json:"lon"`
}

func HandleGetGamePosition(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	round, err := strconv.Atoi(request.PathParameters["round"])
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 404)
	}
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}

	rounds := game.GamesRounds
	if len(rounds) < round {
		return pkg.GenerateResponse("no more rounds", 400)
	}
	r := rounds[round-1]

	bytes, err := json.Marshal(PositionResponse{
		Areas:  game.Areas,
		PanoID: r.PanoID,
		Lat:    r.StartPosition.Lat,
		Lon:    r.StartPosition.Lon,
	})
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

func HandleGetGameStats(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	game, getGameError := getGame(request)
	if getGameError != nil {
		return getGameError.response
	}
	bytes, err := json.Marshal(game)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}

type GetGameError struct {
	error
	response events.APIGatewayProxyResponse
}

func getGame(request events.APIGatewayProxyRequest) (*pkg.Room, *GetGameError) {
	gameID := request.PathParameters["gameID"]

	if gameID == "" {
		return nil, &GetGameError{
			error:    errors.New("no game id given"),
			response: pkg.GenerateResponse("no game id given", 400),
		}
	}
	game, err := pkg.GetRoom(gameID)
	if err != nil {
		return nil, &GetGameError{
			error:    err,
			response: pkg.GenerateResponse(fmt.Sprintf("%v", err), 404),
		}
	}
	return game, nil
}
