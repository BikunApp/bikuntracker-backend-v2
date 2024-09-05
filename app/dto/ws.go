package dto

type CoordinateBroadcastMessage struct {
	Coordinates       []BusCoordinate `json:"coordinates"`
	OperationalStatus int             `json:"operationalStatus"`
}
