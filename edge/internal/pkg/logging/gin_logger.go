package logging

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"time"
)

func GinLogger(skipPaths ...string) gin.HandlerFunc {
	var skip map[string]struct{}

	if length := len(skipPaths); length > 0 {
		skip = make(map[string]struct{}, length)

		for _, path := range skipPaths {
			skip[path] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		rawQuery := c.Request.URL.RawQuery

		// Process request
		c.Next()

		latency := time.Since(start)

		hostname, err := os.Hostname()
		if err != nil {
			hostname = "unknown"
		}

		// Log only when path is not being skipped
		if _, ok := skip[path]; ok {
			return
		}

		statusCode := c.Writer.Status()

		fields := log.Fields{
			"hostname":   hostname,
			"statusCode": statusCode,
			"latency":    latency.String(),
			"clientIP":   c.ClientIP(),
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"rawQuery":   rawQuery,
			"userAgent":  c.Request.UserAgent(),
			"referer":    c.Request.Referer(),
			"dataLength": c.Writer.Size(),
		}

		entry := log.WithFields(fields)

		if len(c.Errors) > 0 {
			errMsg := c.Errors.ByType(gin.ErrorTypePrivate).String()
			entry.Error(errMsg)
		} else if statusCode >= http.StatusInternalServerError {
			entry.Error()
		} else if statusCode >= http.StatusBadRequest {
			entry.Warn()
		} else {
			entry.Info()
		}
	}
}
