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
}

type HorizonAPIService struct {
	service     *echo.Echo
	serverPort  int
	metricsPort int
	clientURL   string
	clientName  string
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
	return HorizonAPIService{
		service: service,

		serverPort:  serverPort,
		metricsPort: metricsPort,
		clientURL:   clientURL,
		clientName:  clientName,
	}
}

// Client implements APIService.
func (h HorizonAPIService) Client() *echo.Echo {
	panic("unimplemented")
}

// Run implements APIService.
func (h HorizonAPIService) Run(ctx context.Context) error {
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
func (h HorizonAPIService) Stop(ctx context.Context) error {
	if err := h.service.Shutdown(ctx); err != nil {
		return eris.New("failed to gracefully shutdown server")
	}
	return nil
}

// package horizon

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"regexp"
// 	"strings"
// 	"time"

// 	"github.com/labstack/echo-contrib/echoprometheus"
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// 	"go.uber.org/zap"
// 	"golang.org/x/time/rate"

// 	echoSwagger "github.com/swaggo/echo-swagger"
// )

// type HorizonSubscribeMessage struct {
// 	Action string `json:"action"`
// 	Topic  string `json:"topic"`
// }

// type HorizonRequest struct {
// 	service *echo.Echo
// 	config  *HorizonConfig
// 	log     *HorizonLog
// }

// // Compile a regular expression to match suspicious paths
// var suspiciousPathPattern = regexp.MustCompile(`(?i)\.(env|yaml|yml|ini|config|conf|xml|git|htaccess|htpasswd|backup|secret|credential|password|private|key|token|dump|database|db|logs|debug)$|dockerfile|Dockerfile`)

// func NewHorizonRequest(
// 	config *HorizonConfig,
// 	log *HorizonLog,
// ) (*HorizonRequest, error) {
// 	e := echo.New()

// 	// 1. Pre-middleware: normalize trailing slashes
// 	e.Pre(middleware.RemoveTrailingSlash())
// 	var csp []string
// 	csp = append(csp,
// 		"default-src 'self'",
// 		"img-src 'self' data:",
// 		"font-src 'self'",
// 		"connect-src 'self' ws://localhost:3000", // or wss:// if using HTTPS
// 		"object-src 'none'",
// 		"base-uri 'self'",
// 		"frame-ancestors 'none'",
// 	)

// 	if config.CanDebug() {
// 		csp = append(csp,
// 			"script-src 'self' 'unsafe-inline'",
// 			"style-src 'self' 'unsafe-inline'",
// 		)
// 	} else {
// 		csp = append(csp,
// 			"script-src 'self'",
// 			"style-src 'self'",
// 		)
// 	}
// 	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
// 		// XSS filter in modern browsers
// 		XSSProtection: "1; mode=block",
// 		// Prevent MIME type sniffing for scripts/styles
// 		ContentTypeNosniff: "nosniff",
// 		// Prevent this site from being framed
// 		XFrameOptions: "DENY",
// 		// Strict Transport Security
// 		//   max-age = 1 year, include subdomains, preload flag
// 		HSTSMaxAge:            31536000, // seconds
// 		HSTSExcludeSubdomains: true,
// 		HSTSPreloadEnabled:    true,
// 		// Prevent browsers from sending a Referer header except to same-origin
// 		// or entirely no-referrer if you prefer.
// 		ReferrerPolicy: "strict-origin-when-cross-origin",
// 		// Disable powerful features you aren’t using
// 		// A tight CSP: only allow self for scripts/styles/etc,
// 		// no inline scripts or eval, allow data: for images if you use SVG/data URIs
// 		ContentSecurityPolicy: strings.Join(csp, "; "),
// 	}))

// 	// 3. Request logging (capture all incoming requests)
// 	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
// 		LogURI:           true,
// 		LogURIPath:       true,
// 		LogStatus:        true,
// 		LogMethod:        true,
// 		LogRemoteIP:      true,
// 		LogHost:          true,
// 		LogUserAgent:     true,
// 		LogReferer:       true,
// 		LogLatency:       true,
// 		LogRequestID:     true,
// 		LogContentLength: true,
// 		LogResponseSize:  true,
// 		LogHeaders:       []string{"Authorization", "Content-Type"},
// 		LogQueryParams:   []string{"*"},
// 		LogFormValues:    []string{"*"},

// 		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
// 			log.Log(LogEntry{
// 				Category: CategoryRequest,
// 				Level:    LevelInfo,
// 				Message: fmt.Sprintf("[Request] %s %s | Status: %d | IP: %s | UA: %s | Referer: %s | Latency: %s | Size: %d",
// 					v.Method,
// 					v.URI,
// 					v.Status,
// 					v.RemoteIP,
// 					v.UserAgent,
// 					v.Referer,
// 					v.Latency,
// 					v.ResponseSize,
// 				),
// 				Fields: []zap.Field{
// 					zap.String("method", v.Method),
// 					zap.String("uri", v.URI),
// 					zap.Int("status", v.Status),
// 					zap.String("remote_ip", v.RemoteIP),
// 					zap.String("host", v.Host),
// 					zap.String("user_agent", v.UserAgent),
// 					zap.String("referer", v.Referer),
// 					zap.String("request_id", v.RequestID),
// 					zap.String("content_length", v.ContentLength),
// 					zap.Int64("response_size", v.ResponseSize),
// 					zap.Any("headers", v.Headers),
// 					zap.Any("query_params", v.QueryParams),
// 					zap.Any("form_values", v.FormValues),
// 				},
// 			})
// 			return nil
// 		},
// 	}))

// 	// 5. Rate limiting
// 	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(rate.Limit(20))))

// 	// 7. Custom path inspection (suspicious files & .well-known)
// 	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			path := strings.ToLower(c.Request().URL.Path)
// 			if suspiciousPattern := suspiciousPathPattern.MatchString(path); suspiciousPattern {
// 				log.Log(LogEntry{
// 					Category: CategoryHijack,
// 					Level:    LevelWarn,
// 					Message:  fmt.Sprintf("Suspicious path accessed: %s", path),
// 					Fields: []zap.Field{
// 						zap.String("method", c.Request().Method),
// 						zap.String("remote_ip", c.Request().RemoteAddr),
// 						zap.String("user_agent", c.Request().UserAgent()),
// 						zap.String("uri", c.Request().RequestURI),
// 						zap.String("host", c.Request().Host),
// 						zap.String("referer", c.Request().Referer()),
// 						zap.String("path", path),
// 						zap.String("request_id", c.Request().Header.Get(echo.HeaderXRequestID)),
// 						zap.String("query_params", c.QueryString()),
// 						zap.String("body", GetRequestBody(c)),
// 					},
// 				})
// 				return c.String(http.StatusForbidden, "Access forbidden")
// 			}

// 			if strings.HasPrefix(path, "/.well-known/") {
// 				return c.String(http.StatusNotFound, "Path not found")
// 			}
// 			return next(c)
// 		}
// 	})

// 	// 8. CORS
// 	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
// 		AllowOrigins: []string{
// 			"http://0.0.0.0",
// 			"http://0.0.0.0:80",
// 			"http://0.0.0.0:3000",
// 			"http://0.0.0.0:3001",
// 			"http://0.0.0.0:4173",
// 			"http://0.0.0.0:8080",

// 			// Client Docker
// 			"http://client",
// 			"http://client:80",
// 			"http://client:3000",
// 			"http://client:3001",
// 			"http://client:4173",
// 			"http://client:8080",

// 			// Localhost
// 			"http://localhost",
// 			"http://localhost:80",
// 			"http://localhost:3000",
// 			"http://localhost:3001",
// 			"http://localhost:4173",
// 			"http://localhost:8080",
// 			"http://localhost:5173",
// 			"http://localhost:5174",
// 			"http://localhost:5175",
// 			config.AppClientURL,
// 		},
// 		AllowMethods: []string{
// 			http.MethodGet,
// 			http.MethodPost,
// 			http.MethodPatch,
// 			http.MethodPut,
// 			http.MethodDelete,
// 			http.MethodOptions,
// 		}, AllowHeaders: []string{
// 			echo.HeaderOrigin,
// 			echo.HeaderContentType,
// 			echo.HeaderAccept,
// 			echo.HeaderAuthorization,
// 			echo.HeaderXRequestedWith,
// 		}, ExposeHeaders: []string{echo.HeaderContentLength},
// 		AllowCredentials: true, // must be true if the client sends cookies/auth
// 		MaxAge:           3600,
// 	}))

// 	// 9. Metrics middleware
// 	e.Use(echoprometheus.NewMiddleware(config.AppName))

// 	e.GET("/health", func(c echo.Context) error {
// 		return c.String(200, "OK")
// 	})
// 	return &HorizonRequest{
// 		service: e,
// 		config:  config,
// 		log:     log,
// 	}, nil
// }

// func (hr *HorizonRequest) Run(routes ...func(*echo.Echo)) error {

// 	for _, r := range routes {
// 		r(hr.service)
// 	}

// 	go func() {
// 		metrics := echo.New()
// 		metrics.GET("/metrics", echoprometheus.NewHandler())
// 		if err := metrics.Start(fmt.Sprintf(":%d", hr.config.AppMetricsPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
// 			// skip
// 		}
// 	}()

// 	go func() {
// 		hr.service.GET("/swagger/*", echoSwagger.WrapHandler)
// 		hr.service.Logger.Fatal(hr.service.Start(
// 			fmt.Sprintf(":%d", hr.config.AppPort),
// 		))
// 	}()

// 	return nil
// }

// func (hr *HorizonRequest) Stop() error {
// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()
// 	if err := hr.service.Shutdown(ctx); err != nil {
// 		return fmt.Errorf("failed to gracefully shutdown server: %w", err)
// 	}
// 	return nil
// }

// func (hr *HorizonRequest) Service() *echo.Echo {
// 	return hr.service
// }
