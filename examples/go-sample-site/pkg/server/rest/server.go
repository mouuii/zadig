/*
Copyright 2021 The KodeRover Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rest

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go-sample-site/pkg/log"
	ginmiddleware "go-sample-site/pkg/middleware/gin"
	"go-sample-site/pkg/util/ginzap"
	"go-sample-site/version"
)

type engine struct {
	*gin.Engine

	mode string
}

func NewEngine() *engine {
	s := &engine{mode: gin.DebugMode}

	gin.SetMode(s.mode)

	s.injectMiddlewares()
	s.injectRouters()

	return s
}

func (s *engine) injectMiddlewares() {
	g := gin.New()
	defer func() {
		s.Engine = g
	}()

	if s.mode == gin.TestMode {
		return
	}

	g.Use(ginmiddleware.RequestLog(log.NewFileLogger("/tmp/requests.log")))
	g.Use(gin.Recovery())
	g.Use(ginmiddleware.RequestID())
}

func (s *engine) injectRouters() {
	g := s.Engine

	g.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "Invalid path: %s", c.Request.URL.Path)
	})
	g.HandleMethodNotAllowed = true
	g.NoMethod(func(c *gin.Context) {
		c.String(http.StatusMethodNotAllowed, "Method not allowed: %s %s", c.Request.Method, c.Request.URL.Path)
	})

	apiRouters := g.Group("")
	apiRouters.GET("", func(c *gin.Context) {
		ginzap.WithContext(c).Debug("测试日志")
		time.Sleep(time.Millisecond)

		c.String(200, fmt.Sprintf(
			"Version:\t%s\nBuild Number:\t%s\nGit Commit:\t%s",
			version.Version, version.BuildNumber, version.GitCommit,
		))
	})

	s.Engine = g
}

