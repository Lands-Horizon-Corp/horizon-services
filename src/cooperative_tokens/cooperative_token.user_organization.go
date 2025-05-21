package cooperative_tokens

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type UserOrganizatonClaim struct {
	UserOrganizatonID string `json:"user_organization_id"`
	UserID            string `json:"user_id"`
	BranchID          string `json:"branch_id"`
	OrganizationID    string `json:"organization_id"`
	jwt.RegisteredClaims
}

func (c UserOrganizatonClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserOrganizatonToken struct {
	Token *horizon.HorizonTokenService[UserOrganizatonClaim]
}

func NewUserOrganizatonToken(provider *src.Provider) *UserOrganizatonToken {
	service := &horizon.HorizonTokenService[UserOrganizatonClaim]{
		Name:   provider.Service.Environment.GetString("APP_NAME", ""),
		Secret: provider.Service.Environment.GetByteSlice("APP_TOKEN", ""),
	}
	return &UserOrganizatonToken{Token: service}
}
