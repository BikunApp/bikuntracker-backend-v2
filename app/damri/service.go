package damri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
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

func (s *service) Authenticate() (token string, err error) {
	authData, err := json.Marshal(dto.DamriAuthRequestBody{
		Username: s.config.DamriLoginUsername,
		Password: s.config.DamriLoginPassword,
	})
	if err != nil {
		err = fmt.Errorf("unable to marshal damri auth credentials: %s", err.Error())
		return
	}

	resp, err := http.Post(s.config.DamriApi+"/auth", "application/json", bytes.NewBuffer(authData))
	if err != nil {
		err = fmt.Errorf("something went wrong when doing POST to Damri API: %s", err.Error())
		return
	}

	body, err := utils.ParseResponseBody[dto.DamriAuthResponse](resp)
	if err != nil {
		return
	}

	token = body.Data.Token
	return
}

func (s *service) GetAllBusStatus() (res []dto.BusStatus, err error) {
	request, err := http.NewRequest("GET", s.config.BikunAdminApi+"/bus/status", nil)
	request.Header.Set("api_key", s.config.BikunAdminApiKey)
	if err != nil {
		err = fmt.Errorf("unable to create request: %s", err.Error())
		return
	}

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("unable to execute HTTP request to fetch bus status: %s", err.Error())
		return
	}

	body, err := utils.ParseResponseBody[dto.BikunAdminGetAllBusStatusResponse](resp)
	if err != nil {
		return
	}

	res = body.Data
	return
}

func (s *service) GetBusCoordinates(imeiList []string) (res []dto.BusCoordinate, err error) {
	body, err := json.Marshal(dto.DamriGetCoordinatesRequestBody{
		Imei: imeiList,
	})
	if err != nil {
		err = fmt.Errorf("unable to marshal imeiList: %s", err.Error())
		return
	}

	request, err := http.NewRequest("POST", s.config.DamriApi+"/tg_coordinate", bytes.NewBuffer(body))
	if err != nil {
		err = fmt.Errorf("unable to create request: %s", err.Error())
		return
	}

	request.Header.Set("Authorization", "Bearer "+s.config.Token)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		err = fmt.Errorf("unable to execute HTTP request to fetch bus coordinates: %s", err.Error())
		return
	}

	respBody, err := utils.ParseResponseBody[dto.DamriGetCoordinatesResponse](resp)
	if err != nil {
		return
	}

	res = respBody.Data
	return
}
