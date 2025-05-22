package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/services/horizon"
	"github.com/lands-horizon/horizon-server/src"
)

type Controller struct {
	provider *src.Provider
}

func NewController(provider *src.Provider) (*Controller, error) {
	return &Controller{provider: provider}, nil
}

func (c *Controller) Routes() {
	req := c.provider.Service.Request

	req.RegisterRoute(horizon.Route{
		Route:    "/health",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/health",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/health",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/health",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/health",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})

	req.RegisterRoute(horizon.Route{
		Route:    "/sure",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/sure",
		Method:   "GET",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/sure",
		Method:   "DELETE",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/sure",
		Method:   "PUT",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
	req.RegisterRoute(horizon.Route{
		Route:    "/sure",
		Method:   "POST",
		Request:  "",
		Response: "string", // or "OK"
		Note:     "Health check endpoint",
	}, func(c echo.Context) error {
		return c.String(200, "OK")
	})
}
