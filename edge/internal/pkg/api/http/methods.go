package http

import (
	"devais.it/kronos/internal/pkg/config"
	"devais.it/kronos/internal/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

type methods struct {
	conf   *config.HTTPConfig
	router *gin.RouterGroup
}

func (m *methods) replyCreatedData() bool {
	return m.conf.ReplyCreatedData
}

func (m *methods) writeError(c *gin.Context, code int, err error) {
	errBody := gin.H{
		"error": eris.ToString(err, false),
	}
	c.JSON(code, errBody)
}

// writeServiceError converts a generic error coming from underlying services
// to the corresponding HTTP error and renders it to Gin context
func (m *methods) writeServiceError(c *gin.Context, err error) {
	if eris.Is(err, gorm.ErrRecordNotFound) {
		m.writeError(c, http.StatusNotFound, err)
		return
	}

	if eris.Is(err, gorm.ErrInvalidData) ||
		eris.Is(err, gorm.ErrInvalidField) ||
		eris.Is(err, db.ErrMissingID) ||
		eris.Is(err, db.ErrInvalidPagination) {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	if strings.Contains(eris.ToString(err, true), "UNIQUE constraint failed") {
		m.writeError(c, http.StatusBadRequest, err)
		return
	}

	m.writeError(c, http.StatusInternalServerError, err)
}
