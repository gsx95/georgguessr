package main

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/aws/aws-lambda-go/events"
)

func main() {

	methods := pkg.MethodHandlers{
		"/countries/{continent}": {
			GET: HandleGetCountriesInContinent,
		},
	}

	pkg.StartLambda(methods)
}

func HandleGetCountriesInContinent(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	continentCode := req.PathParameters["continent"]

	countries, err := pkg.GetCountries(continentCode)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 400)
	}
	bytes, err := json.Marshal(countries)
	if err != nil {
		return pkg.GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return pkg.GenerateResponse(string(bytes), 200)
}
