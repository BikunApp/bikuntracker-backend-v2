package dto

import (
	"encoding/xml"

	"github.com/FreeJ1nG/bikuntracker-backend/app/models"
)

type SSOLoginRequestBody struct {
	Ticket  string `json:"ticket"`
	Service string `json:"service"`
}

type VerifyTicketResponse struct {
	XMLName               xml.Name              `xml:"http://www.yale.edu/tp/cas serviceResponse"`
	AuthenticationSuccess AuthenticationSuccess `xml:"authenticationSuccess"`
}

type AuthenticationSuccess struct {
	User       string     `xml:"user"`
	Attributes Attributes `xml:"attributes"`
}

type Attributes struct {
	LDAPCn    string `xml:"ldap_cn"`
	KdOrg     string `xml:"kd_org"`
	PeranUser string `xml:"peran_user"`
	Nama      string `xml:"nama"`
	NPM       string `xml:"npm"`
}

type TokenResponse struct {
	AccessToken  string `json:"access"`
	RefreshToken string `json:"refresh"`
}

type SSOLoginResponse struct {
	TokenResponse
	User models.User `json:"user"`
}

type RefreshTokenRequestBody struct {
	RefreshToken string `json:"refresh"`
}
