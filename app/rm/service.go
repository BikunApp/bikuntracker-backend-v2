package rm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

type service struct {
	config *utils.Config
}

func NewService(config *utils.Config) *service {
	return &service{
		config: config,
	}
}

func (s *service) DetectLane(data map[string][]*models.BusCoordinate) (res dto.DetectRouteResponse, err error) {
	currentPoints := make(map[string][]dto.Point)
	for imei, points := range data {
		formattedPoints := make([]dto.Point, 0)
		for _, point := range points {
			formattedPoints = append(formattedPoints, dto.Point{
				TimeStamp: point.GpsTime.Unix(),
				Lat:       point.Latitude,
				Lng:       point.Longitude,
			})
		}
		currentPoints[imei] = formattedPoints
	}

	body, err := json.Marshal(dto.DetectRouteRequestBody{
		CurrentPoints: currentPoints,
	})
	if err != nil {
		err = fmt.Errorf("unable to marshal detectRouteRequestBody: %w", err)
		return
	}

	request, err := http.NewRequest("POST", s.config.RMApi+"/detect-route/", bytes.NewBuffer(body))
	if err != nil {
		err = fmt.Errorf("unable to create request: %w", err)
		return
	}

	request.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("unable to execute HTTP request to detect route: %w", err)
		return
	}

	res, err = utils.ParseResponseBody[dto.DetectRouteResponse](resp)
	if err != nil {
		return
	}

	return
}
