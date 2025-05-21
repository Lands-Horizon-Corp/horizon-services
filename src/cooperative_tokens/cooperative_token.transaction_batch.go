package cooperative_tokens

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type TransactionBatchClaim struct {
	TransactionBatchID string `json:"transaction_batch_id"`
	BranchID           string `json:"branch_id"`
	OrganizationID     string `json:"organization_id"`
	UserID             string `json:"user_id"`
	jwt.RegisteredClaims
}

func (c TransactionBatchClaim) GetRegisteredClaims() *jwt.RegisteredClaims {
	return &c.RegisteredClaims
}

type TransactionBatchToken struct {
	Token *horizon.HorizonTokenService[TransactionBatchClaim]
}

func NewTransactionBatchToken(provider *src.Provider) *TransactionBatchToken {
	service := &horizon.HorizonTokenService[TransactionBatchClaim]{
		Name:   provider.Service.Environment.GetString("APP_NAME", ""),
		Secret: provider.Service.Environment.GetByteSlice("APP_TOKEN", ""),
	}
	return &TransactionBatchToken{Token: service}
}
