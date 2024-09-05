package damri

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/FreeJ1nG/bikuntracker-backend/utils"
)

const (
	MORNING_ROUTE   = 0
	NORMAL_ROUTE    = 1
	NOT_OPERATIONAL = 2
)

type service struct {
	config *utils.Config
	util   interfaces.DamriUtil
}

func NewService(config *utils.Config, util interfaces.DamriUtil) *service {
	return &service{
		config: config,
		util:   util,
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

	resp, err := http.DefaultClient.Do(request)
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
	resp, err := http.DefaultClient.Do(request)
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

func (s *service) GetOperationalStatus() (int, error) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		err = fmt.Errorf("unable to load Asia/Jakarta location")
		return NOT_OPERATIONAL, err
	}

	now := time.Now().In(loc)
	dayOfWeek := int(now.Weekday())
	currentTime := now.Hour()*60 + now.Minute()

	// If day is Monday - Friday
	if dayOfWeek >= 1 && dayOfWeek <= 5 {
		if currentTime >= s.util.GetHMInMinutes(6, 50) && currentTime < s.util.GetHMInMinutes(9, 0) {
			return MORNING_ROUTE, nil
		} else if currentTime >= s.util.GetHMInMinutes(9, 0) && currentTime < s.util.GetHMInMinutes(21, 0) {
			return NORMAL_ROUTE, nil
		} else {
			return NOT_OPERATIONAL, nil
		}
	}

	// If day is Saturday
	if dayOfWeek == 6 {
		if currentTime >= s.util.GetHMInMinutes(6, 50) && currentTime < s.util.GetHMInMinutes(15, 30) {
			return NORMAL_ROUTE, nil
		} else {
			return NOT_OPERATIONAL, nil
		}
	}

	// Sunday means that Bikun is not operational
	return NOT_OPERATIONAL, nil
}
