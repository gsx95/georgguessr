package main

import "georgguessr.com/lambda-room/creation"

type ExistsResponse struct {
	Exists bool `json:"exists"`
}

type RoomCreationResponse struct {
	RoomId string `json:"roomId"`
	GenPositions creation.Positions `json:"generatedPositions"`
}