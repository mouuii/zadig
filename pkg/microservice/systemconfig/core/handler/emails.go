package handler

import (
	"github.com/gin-gonic/gin"

	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
)

func AddHost(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	ctx.Resp = &feature{
		Name: "xx",
	}
}
