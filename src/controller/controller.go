package controller

import (
	"github.com/labstack/echo/v4"
	"github.com/lands-horizon/horizon-server/src"
)

type Routes struct {
	route    string
	request  string
	response string
	method   string
}

type Controller struct {
	provider   *src.Provider
	RoutesList []Routes
}

func NewController(provider *src.Provider) (*Controller, error) {
	return &Controller{provider: provider, RoutesList: []Routes{}}, nil
}

func (c *Controller) Routes() {
	c.RegisterRouteGET("No content", "OK", "/coop", func(c echo.Context) error {
		return c.String(200, "OK")
	})
}

func (c *Controller) RegisterRouteGET(request string, response string, route string, h func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	server := c.provider.Service.Request.Client()
	server.GET(route, h, m...)
	c.RoutesList = append(c.RoutesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "GET",
	})
}

func (c *Controller) RegisterRoutePOST(request string, response string, route string, h func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	server := c.provider.Service.Request.Client()
	server.POST(route, h, m...)
}

func (c *Controller) RegisterRouteDELETE(request string, response string, route string, h func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	server := c.provider.Service.Request.Client()
	server.DELETE(route, h, m...)
}

func (c *Controller) RegisterRoutePUT(request string, response string, route string, h func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	server := c.provider.Service.Request.Client()
	server.PUT(route, h, m...)
}
