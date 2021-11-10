package handler

import (
	"github.com/gin-gonic/gin"
)

type Router struct{}

func (*Router) Inject(router *gin.RouterGroup) {

	dev := router.Group("")
	{
		dev.POST("/workflow/workflowTask/create", CreateWorkflowTask)
		dev.DELETE("/workflow/workflowTask/id/:id/pipelines/:name/cancel", CancelWorkflowTask)
		dev.POST("/workflow/workflowTask/id/:id/pipelines/:name/restart", RestartWorkflowTask)
		dev.GET("/workflow/workflowTask", ListWorkflowTask)
		dev.GET("/dc/releases", ListDelivery)
	}
}
