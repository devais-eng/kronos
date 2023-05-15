package http

import (
	"context"
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/health"
	"devais.it/kronos/internal/pkg/logging"
	"github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Server struct {
	conf   *config.HTTPConfig
	engine *gin.Engine
	server *http.Server
}

func SetMode(debugMode bool) {
	if debugMode {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
}

// NewServer creates a new HTTP server.
// The server can be started with the Server.Start method
func NewServer(conf *config.HTTPConfig) *Server {
	SetMode(conf.DebugMode)

	engine := gin.New()

	engine.Use(logging.GinLogger())

	if conf.Sentry.Enabled {
		options := sentrygin.Options{
			Repanic:         true,
			WaitForDelivery: conf.Sentry.WaitForDelivery,
			Timeout:         conf.Sentry.DeliveryTimeout,
		}
		engine.Use(sentrygin.New(options))
	}

	server := &http.Server{
		Addr:    conf.Address(),
		Handler: engine,
	}

	newItemMethods(engine, conf, "/")
	newRelationMethods(engine, conf, "/")
	newAttributeMethods(engine, conf, "/")
	newEventMethods(engine, conf, "/")

	engine.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	if conf.PprofEnabled {
		pprof.Register(engine, "pprof")
	}

	if conf.DebugMode {
		engine.GET("/testPanic", func(c *gin.Context) {
			panic("Test panic")
		})
	}

	engine.GET("/health", func(c *gin.Context) {
		result := health.Check()

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, result)
		} else {
			c.String(http.StatusOK, "ok")
		}
	})

	return &Server{
		conf:   conf,
		engine: engine,
		server: server,
	}
}

func (s *Server) Start() {
	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			if eris.Is(err, http.ErrServerClosed) {
				log.Info("HTTP server closed")
			} else {
				logging.Error(err, "HTTP server error")
			}
		}
	}()
}

func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.conf.Timeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}
