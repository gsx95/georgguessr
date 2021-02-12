package handler

import (
	"backend/data"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
)

func HandleGetCountriesInContinent(req events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	continentCode := req.PathParameters["continent"]

	countries, err := data.GetCountries(continentCode)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 400)
	}
	bytes, err := json.Marshal(countries)
	if err != nil {
		return GenerateResponse(fmt.Sprintf("%v", err), 500)
	}
	return GenerateResponse(string(bytes), 200)
}
