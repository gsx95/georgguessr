package main

import "georgguessr.com/pkg"

type positionResponse struct {
	Areas  [][]pkg.GeoPoint `json:"areas,omitempty"`
	PanoID string           `json:"panoId"`
	Lat    float64          `json:"lat"`
	Lon    float64          `json:"lon"`
}

type playersResponse struct {
	Players []string `json:"players"`
}
