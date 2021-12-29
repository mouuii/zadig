package gin

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const timeISO8601 = "2006-01-02T15:04:05.000Z0700"

func RequestLog(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		var body []byte
		// request body is a ReadCloser, it can be read only once.
		if c.Request != nil && c.Request.Body != nil {
			var buf bytes.Buffer
			tee := io.TeeReader(c.Request.Body, &buf)
			body, _ = ioutil.ReadAll(tee)
			c.Request.Body = ioutil.NopCloser(&buf)
		}

		// Process request
		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		logger.Info("",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.ByteString("body", body),
			zap.Int("size", c.Writer.Size()),
			zap.String("clientIP", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("start", start.Format(timeISO8601)),
			zap.Duration("latency", latency),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}
