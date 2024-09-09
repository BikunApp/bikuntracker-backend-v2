package auth

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/FreeJ1nG/bikuntracker-backend/app/dto"
	"github.com/FreeJ1nG/bikuntracker-backend/app/interfaces"
	"github.com/golang-jwt/jwt/v5"
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

func (s *service) SsoLogin(ticket, serviceName string) (res dto.SSOLoginResponse, status int, err error) {
	status = http.StatusOK

	request, err := http.NewRequest("GET", fmt.Sprintf("https://sso.ui.ac.id/cas2/serviceValidate?ticket=%s&service=%s", ticket, serviceName), nil)
	if err != nil {
		err = fmt.Errorf("unable to create HTTP request: %w", err)
		status = http.StatusInternalServerError
		return
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		err = fmt.Errorf("unable to execute SSO request: %w", err)
		status = http.StatusInternalServerError
		return
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		err = fmt.Errorf("unable to read response body: %w", err)
		status = http.StatusInternalServerError
		return
	}

	var credentials dto.VerifyTicketResponse
	err = xml.Unmarshal(body, &credentials)
	if err != nil {
		err = fmt.Errorf("unable to unmarshal SSO XML: %w", err)
		status = http.StatusInternalServerError
		return
	}

	email := credentials.AuthenticationSuccess.User + "@ui.ac.id"

	user, err := s.repo.GetOrCreateUser(
		credentials.AuthenticationSuccess.Attributes.Nama,
		credentials.AuthenticationSuccess.Attributes.NPM,
		email,
	)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	res.User = *user

	generatedAccessToken, generatedRefreshToken, err := s.util.GenerateTokenPair(user)
	if err != nil {
		status = http.StatusInternalServerError
		return
	}

	res.AccessToken = generatedAccessToken
	res.RefreshToken = generatedRefreshToken

	return
}

func (s *service) RefreshToken(refreshToken string) (res *dto.TokenResponse, status int, err error) {
	status = http.StatusOK

	rt, err := s.util.ToJwtToken(refreshToken)
	if err != nil {
		err = fmt.Errorf("unable to decode refresh token: %w", err)
		status = http.StatusBadRequest
		return
	}

	tokenClaims, ok := rt.Claims.(jwt.MapClaims)
	if !ok {
		err = fmt.Errorf("unable to get token claims: %w", err)
		status = http.StatusBadRequest
		return
	}

	user, err := s.repo.GetUserByEmail(tokenClaims["sub"].(string))
	if err != nil {
		err = fmt.Errorf("unable to get user: %w", err)
		status = http.StatusNotFound
		return
	}

	accessToken, refreshToken, err := s.util.GenerateTokenPair(user)
	if err != nil {
		err = fmt.Errorf("unable to generate new token: %w", err)
		status = http.StatusInternalServerError
		return
	}

	res = &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	return
}
