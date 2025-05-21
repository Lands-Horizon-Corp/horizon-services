package main

import (
	"github.com/lands-horizon/horizon-server/src"
	"github.com/lands-horizon/horizon-server/src/cooperative_tokens"
	"go.uber.org/fx"
)

func main() {
	app := fx.New(
		fx.Provide(
			src.NewProvider,
			cooperative_tokens.NewUserToken,
			cooperative_tokens.NewTransactionBatchToken,
			cooperative_tokens.NewUserOrganizatonToken,
		),
		fx.Invoke(
			src.NewProvider,
		),
	)
	app.Run()

}
