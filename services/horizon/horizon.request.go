package horizon

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rotisserie/eris"
	echoSwagger "github.com/swaggo/echo-swagger"
	"golang.org/x/time/rate"
)

type APIService interface {
	Run(ctx context.Context) error

	Stop(ctx context.Context) error

	Client() *echo.Echo

	GetRoutes() []Routes

	RegisterRouteGET(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
	RegisterRoutePOST(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
	RegisterRoutePUT(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
	RegisterRouteDELETE(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
	RegisterRoutePATCH(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc)
}

type Routes struct {
	route    string
	request  string
	response string
	method   string
}

type HorizonAPIService struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string

	routesList []Routes
}

var suspiciousPathPattern = regexp.MustCompile(`(?i)\.(env|yaml|yml|ini|config|conf|xml|git|htaccess|htpasswd|backup|secret|credential|password|private|key|token|dump|database|db|logs|debug)$|dockerfile|Dockerfile`)

func NewHorizonAPIService(
	serverPort int,
	metricsPort int,
	clientURL string,
	clientName string,
) APIService {
	service := echo.New()

	service.Pre(middleware.RemoveTrailingSlash())

	service.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		HSTSExcludeSubdomains: true,
		HSTSPreloadEnabled:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
	}))

	service.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:           true,
		LogURIPath:       true,
		LogStatus:        true,
		LogMethod:        true,
		LogRemoteIP:      true,
		LogHost:          true,
		LogUserAgent:     true,
		LogReferer:       true,
		LogLatency:       true,
		LogRequestID:     true,
		LogContentLength: true,
		LogResponseSize:  true,
		LogHeaders:       []string{"Authorization", "Content-Type"},
		LogQueryParams:   []string{"*"},
		LogFormValues:    []string{"*"},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			return nil
		},
	}))

	// 5. Rate limiting
	service.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

	service.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := strings.ToLower(c.Request().URL.Path)
			if suspiciousPattern := suspiciousPathPattern.MatchString(path); suspiciousPattern {
				return c.String(http.StatusForbidden, "Access forbidden")
			}
			if strings.HasPrefix(path, "/.well-known/") {
				return c.String(http.StatusNotFound, "Path not found")
			}
			return next(c)
		}
	})

	service.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://0.0.0.0",
			"http://0.0.0.0:80",
			"http://0.0.0.0:3000",
			"http://0.0.0.0:3001",
			"http://0.0.0.0:4173",
			"http://0.0.0.0:8080",

			// Client Docker
			"http://client",
			"http://client:80",
			"http://client:3000",
			"http://client:3001",
			"http://client:4173",
			"http://client:8080",

			// Localhost
			"http://localhost",
			"http://localhost:80",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:4173",
			"http://localhost:8080",
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:5175",
			clientURL,
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		}, AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			echo.HeaderXRequestedWith,
		}, ExposeHeaders: []string{echo.HeaderContentLength},
		AllowCredentials: true, // must be true if the client sends cookies/auth
		MaxAge:           3600,
	}))

	// 9. Metrics middleware
	service.Use(echoprometheus.NewMiddleware(clientName))

	service.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})
	return &HorizonAPIService{
		service: service,

		serverPort:  serverPort,
		metricsPort: metricsPort,
		clientURL:   clientURL,
		clientName:  clientName,
		routesList:  []Routes{},
	}
}

// Client implements APIService.
func (h *HorizonAPIService) Client() *echo.Echo {
	return h.service
}

// GetRoutes implements APIService.
func (h *HorizonAPIService) GetRoutes() []Routes {
	return h.routesList
}

// RegisterRouteDELETE implements APIService.
func (h *HorizonAPIService) RegisterRouteDELETE(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	h.service.DELETE(route, callback, m...)
	h.routesList = append(h.routesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "DELETE",
	})
}

// RegisterRouteGET implements APIService.
func (h *HorizonAPIService) RegisterRouteGET(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	h.service.GET(route, callback, m...)
	h.routesList = append(h.routesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "GET",
	})
}

// RegisterRoutePATCH implements APIService.
func (h *HorizonAPIService) RegisterRoutePATCH(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	h.service.PATCH(route, callback, m...)

	h.routesList = append(h.routesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "PATCH",
	})
}

// RegisterRoutePOST implements APIService.
func (h *HorizonAPIService) RegisterRoutePOST(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	h.service.POST(route, callback, m...)
	h.routesList = append(h.routesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "POST",
	})
}

// RegisterRoutePUT implements APIService.
func (h *HorizonAPIService) RegisterRoutePUT(request string, response string, route string, callback func(c echo.Context) error, m ...echo.MiddlewareFunc) {
	h.service.PUT(route, callback, m...)
	h.routesList = append(h.routesList, Routes{
		route:    route,
		request:  request,
		response: response,
		method:   "PUT",
	})
}

// Run implements APIService.
func (h *HorizonAPIService) Run(ctx context.Context) error {
	go func() {
		metrics := echo.New()
		metrics.GET("/metrics", echoprometheus.NewHandler())
		if err := metrics.Start(fmt.Sprintf(":%d", h.metricsPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			// skip
		}
	}()
	go func() {
		h.service.GET("/swagger/*", echoSwagger.WrapHandler)
		h.service.Logger.Fatal(h.service.Start(
			fmt.Sprintf(":%d", h.serverPort),
		))
	}()
	return nil
}

// Stop implements APIService.
func (h *HorizonAPIService) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}
