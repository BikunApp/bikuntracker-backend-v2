package dto

type Point struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	TimeStamp int64   `json:"ts"`
}

type DetectRouteRequestBody struct {
	CurrentPoints map[string][]Point `json:"current_points"`
}

type DetectRouteResponse = map[string]string
