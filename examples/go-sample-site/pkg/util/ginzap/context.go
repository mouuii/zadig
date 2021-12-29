package ginzap

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go-sample-site/pkg/log"
)

const loggerKey = iota

func NewContext(ctx *gin.Context, fields ...zapcore.Field) {
	ctx.Set(strconv.Itoa(loggerKey), WithContext(ctx).With(fields...))
}

func WithContext(ctx *gin.Context) *zap.Logger {
	if ctx == nil {
		return log.Logger()
	}
	l, _ := ctx.Get(strconv.Itoa(loggerKey))
	ctxLogger, ok := l.(*zap.Logger)
	if ok {
		return ctxLogger
	}
	return log.Logger()
}
