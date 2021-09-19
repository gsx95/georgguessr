package creation

import (
	"encoding/json"
	"fmt"
	"georgguessr.com/pkg"
	"github.com/mmcloughlin/spherand"
	"log"
)

type CreatorUnrestricted struct {}

func (cr *CreatorUnrestricted) CreateRoom(reqBody string) (roomId string, genPos *Positions, err error) {
	defer pkg.LogDuration(pkg.Track())
	log.Printf("create unrestricted room %v\n", reqBody)
	room := pkg.Room{}
	err = json.Unmarshal([]byte(reqBody), &room)
	if err != nil {
		return "", nil, pkg.InternalErr(fmt.Sprintf("Error unmarshalling request body: %v %v", reqBody, err))
	}

	var positions Positions

	log.Printf("generate random Positions\n")

	for i := 0; i < room.Rounds + additionalCreationTries; i++ {
		lat, lon := spherand.Geographical()
		positions = append(positions, newPosition(lat, lon))
	}

	log.Printf("generated Positions: %v\n", positions)

	roomId, err = saveRoom(&room)
	if err != nil {
		return "", nil, err
	}
	return roomId, &positions, nil
}