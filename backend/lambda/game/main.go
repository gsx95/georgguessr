package main

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-lambda-go/events"
	"strconv"
	"strings"
)

func main() {

	methods := pkg.MethodHandlers {
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


func HandleGetGuess(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	round, err := getRoundFromRequest(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	scores := game.GamesRounds[round-1].Scores
	return pkg.GenerateResponse(scores, 200)
}

func HandlePostGuess(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	username, err := getStringFromRequest("username", request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	round, err := getRoundFromRequest(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	var guess pkg.Guess

	if err := json.Unmarshal([]byte(request.Body), &guess); err != nil {
		fmt.Println(err)
		return pkg.ErrorResponse(pkg.BadRequestErr("invalid guess body"))
	}

	scores := game.GamesRounds[round-1].Scores
	if _, alreadyExists := scores[username]; alreadyExists {
		return pkg.ErrorResponse(pkg.BadRequestErr("Already posted score for this round"))
	}

	PutGuess(game.ID, username, round-1, guess)

	return pkg.OkResponse()
}

func HandlePostPlayers(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	username, err := getStringFromRequest("username", request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	for _, user := range game.Players {
		if strings.ToLower(user) == strings.ToLower(username) {
			return pkg.ErrorResponse(pkg.BadRequestErr("username already in use"))
		}
	}
	PutUsername(game.ID, username)
	return pkg.OkResponse()
}

func HandleGetPlayers(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	resp := playersResponse{
		game.Players,
	}
	return pkg.GenerateResponse(resp, 200)
}

func HandleGetGameEndResults(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	resp := pkg.Room{
		Rounds:      game.Rounds,
		Players:     game.Players,
		Areas:       game.Areas,
		GamesRounds: game.GamesRounds,
	}

	return pkg.GenerateResponse(resp, 200)
}

func HandleGetGamePosition(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	round, err := getRoundFromRequest(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	rounds := game.GamesRounds
	if len(rounds) < round {
		return pkg.ErrorResponse(pkg.BadRequestErr("Round %d not found for this game"))
	}
	r := rounds[round-1]
	resp := positionResponse{
		Areas:  game.Areas,
		PanoID: r.PanoID,
		Lat:    r.StartPosition.Lat,
		Lon:    r.StartPosition.Lng,
	}
	return pkg.GenerateResponse(resp, 200)
}

func HandleGetGameStats(request events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {
	game, err := getGame(request)
	if err != nil {
		return pkg.ErrorResponse(err)
	}
	return pkg.GenerateResponse(game, 200)
}

func getGame(request events.APIGatewayProxyRequest) (*pkg.Room, error) {
	gameID, err := getStringFromRequest("gameID", request)
	if err != nil {
		return nil, err
	}
	game, err := pkg.GetRoom(gameID)
	if err != nil {
		return nil, err
	}
	return game, nil
}

func getStringFromRequest(key string, request events.APIGatewayProxyRequest) (string, error) {
	value := request.PathParameters[key]
	if value == "" {
		return "", pkg.BadRequestErr(fmt.Sprintf("No %s given in request", key))
	}
	return value, nil
}

func getRoundFromRequest(request events.APIGatewayProxyRequest) (int, error) {
	roundString := request.PathParameters["round"]
	if roundString == "" {
		return 0, pkg.BadRequestErr("No round given in request")
	}
	round, err := strconv.Atoi(roundString)
	if err != nil {
		return 0, pkg.BadRequestErr(fmt.Sprintf("Round must be a number: %s", roundString))
	}
	return round, nil
}