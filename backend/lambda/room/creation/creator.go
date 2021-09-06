package creation

import (
	"fmt"
	"georgguessr.com/pkg"
)

type RoomsCreator interface {
	CreateRoom(requestBody string) (roomID string, err error)
}

func NewCreator(creatorType string) (RoomsCreator, error) {
	switch creatorType {
	case "list":
		return &CreatorPredefinedCities{}, nil
	case "unlimited":
		return &CreatorUnrestricted{}, nil
	case "places":
		return &CreatorUserdefinedCities{}, nil
	case "custom":
		return &CreatorCustomArea{}, nil
	default:
		return nil, pkg.BadRequestErr(fmt.Sprintf("Creation type %s not recognized", creatorType))
	}
}