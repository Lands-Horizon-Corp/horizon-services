package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
)

func (c *Controller) MediaController() {

	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/media",
		Method:   "GET",
		Request:  "",
		Response: "string",
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})

}
