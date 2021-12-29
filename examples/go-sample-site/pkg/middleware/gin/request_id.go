package gin

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-sample-site/pkg/util/ginzap"
)

// RequestID adds a unique request id to the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := uuid.New()
		ginzap.NewContext(c, zap.String("requestID", reqID.String()))

		c.Next()
	}
}
