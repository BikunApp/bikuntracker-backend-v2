package rm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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

func (s *service) DetectLane(imei string, data []*models.BusCoordinate) (res dto.DetectRouteResponse, err error) {
	formattedPoints := make([]dto.Point, 0)
	for _, point := range data {
		formattedPoints = append(formattedPoints, dto.Point{
			TimeStamp: point.GpsTime.Unix(),
			Lat:       point.Latitude,
			Lng:       point.Longitude,
		})
	}

	body, err := json.Marshal(dto.DetectRouteRequestBody{
		CurrentPoints: map[string][]dto.Point{
			imei: formattedPoints,
		},
	})
	if err != nil {
		err = fmt.Errorf("unable to marshal detectRouteRequestBody: %w", err)
		return
	}

	var prettyJson bytes.Buffer
	err = json.Indent(&prettyJson, body, "", "\t")
	if err != nil {
		err = fmt.Errorf("unable to pretty print json body: %w", err)
		return
	}

	log.Println("Detecting route for:", string(prettyJson.Bytes()))

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
