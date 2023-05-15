package http

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/services"
	"github.com/gin-gonic/gin"
	"net/http"
)

type eventMethods struct {
	methods
}

func (m *eventMethods) count(c *gin.Context) {
	count, err := services.GetEventsCount()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}

func (m *eventMethods) getFirst(c *gin.Context) {
	event, err := services.GetFirstEvent()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func (m *eventMethods) getLast(c *gin.Context) {
	event, err := services.GetLastEvent()
	if err != nil {
		m.writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, event)
}

func newEventMethods(engine *gin.Engine, conf *config.HTTPConfig, rootPath string) *eventMethods {
	g := engine.Group(rootPath)

	m := &eventMethods{
		methods{router: g, conf: conf},
	}

	g.
		GET("/events/first", m.getFirst).
		GET("/events/last", m.getLast).
		GET("/events/count", m.count)

	return m
}
