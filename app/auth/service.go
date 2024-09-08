package auth

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
)

type service struct {
	util interfaces.AuthUtil
	repo interfaces.AuthRepository
}

func NewService(util interfaces.AuthUtil, repo interfaces.AuthRepository) *service {
	return &service{
		util: util,
		repo: repo,
	}
}

func (s *service) SSOLogin(ticket, serviceName string) (accessToken string, refreshToken string, err error) {
	request, err := http.NewRequest("GET", fmt.Sprintf("https://sso.ui.ac.id/cas2/serviceValidate?ticket=%s&service=%s", ticket, serviceName), nil)
	if err != nil {
		err = fmt.Errorf("unable to create HTTP request: %w", err)
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("unable to execute SSO request: %w", err)
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("unable to read response body: %w", err)
		return
	}

	var credentials dto.VerifyTicketResponse
	err = xml.Unmarshal(body, &credentials)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal SSO XML: %w", err)
		return
	}

	email := credentials.AuthenticationSuccess.User + "@ui.ac.id"

	user, err := s.repo.GetOrCreateUser(
		credentials.AuthenticationSuccess.Attributes.Nama,
		credentials.AuthenticationSuccess.Attributes.NPM,
		email,
	)
	if err != nil {
		return
	}

	accessToken, refreshToken, err = s.util.GenerateTokenPair(&user)
	if err != nil {
		return
	}

	return
}
