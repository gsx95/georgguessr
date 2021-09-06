package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/mmcloughlin/spherand"
)

type CreatorUnrestricted struct {}

func (cr *CreatorUnrestricted) CreateRoom(reqBody string) (string, error) {
	fmt.Printf("create unrestricted room %v\n", reqBody)
	room := pkg.Room{}
	err := json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	positions := positions{}

	fmt.Printf("generate random positions\n")

	for i := 0; i < room.Rounds +additionalCreationTries; i++ {
		lat, lon := spherand.Geographical()
		positions.Pos = append(positions.Pos, newRoundPos(i, lat, lon))
	}

	fmt.Printf("created positions: %v\n", positions)

	streetViews, err := getStreetviewPositions(positions, room.Rounds)
	if err != nil {
		return "", err
	}
	addStreetViewToRoom(&room, *streetViews)

	return createRoom(&room)
}