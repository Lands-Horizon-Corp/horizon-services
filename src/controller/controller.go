package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src"
)

type Controller struct {
	provider *src.Provider
}

func NewController(provider *src.Provider) (*Controller, error) {
	server := provider.Service.Request.Client()

	server.GET("/health-1", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	return &Controller{
		provider: provider,
	}, nil
}
