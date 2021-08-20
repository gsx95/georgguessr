package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os/exec"
)

type Positions struct {
	Pos []RoundPosition `json:"pos"`
}

type RoundPosition struct {
	Round int `json:"r"`
	Position Position `json:"p"`
}

func NewRoundPos(round int, lat, lng float64) RoundPosition {
	return RoundPosition{
		Round: round,
		Position: Position{
			Lat: lat,
			Lng: lng,
		},
	}
}

type Position struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type pano struct {
	Round  int    `json:"r"`
	PanoID string `json:"id"`
	Pos Position `json:"location"`

}

type StreetViewIDs struct {
	Panos []pano `json:"panos"`
}


func GetStreetviewPositions(positions Positions, num int) StreetViewIDs {
	posJson, err := json.Marshal(positions)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(posJson))

	url := fmt.Sprintf(`file:///opt/bin/index.html?pos=%s`, string(posJson))

	app := "/opt/bin/phantomjs"
	arg0 := "/opt/bin/script.js"
	arg1 := url

	cmd := exec.Command(app, arg0, arg1)
	stdout, err := cmd.Output()
	if err != nil {
		panic(err)
	}

	allStreetViews := StreetViewIDs{}
	json.Unmarshal(stdout, &allStreetViews)
	okStreetViews := StreetViewIDs{
		Panos: []pano{},
	}

	count := 0

	for _, generatedSV := range allStreetViews.Panos {
		if generatedSV.PanoID != "" {
			okStreetViews.Panos = append(okStreetViews.Panos, pano{
				Round: generatedSV.Round,
				Pos: generatedSV.Pos,
				PanoID: generatedSV.PanoID,
			})
			count++
		}
		if count == num {
			break
		}
	}
	return okStreetViews
}