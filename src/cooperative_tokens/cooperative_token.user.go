package cooperative_tokens

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type UserClaim struct {
	UserID        string `json:"user_id"`
	Email         string `json:"email"`
	ContactNumber string `json:"contact_number"`
	Password      string `json:"password"`
	Username      string `json:"username"`
	jwt.RegisteredClaims
}

func (c UserClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type UserToken struct {
	Token *horizon.HorizonTokenService[UserClaim]
}

func NewUserToken(provider *src.Provider) *UserToken {
	service := &horizon.HorizonTokenService[UserClaim]{
		Name:   provider.Service.Environment.GetString("APP_NAME", ""),
		Secret: provider.Service.Environment.GetByteSlice("APP_TOKEN", ""),
	}
	return &UserToken{Token: service}
}
