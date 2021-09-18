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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/util/wait"

	commonmodels "github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/common/repository/models/template"
	commonservice "github.com/koderover/zadig/pkg/microservice/aslan/core/common/service"
	"github.com/koderover/zadig/pkg/microservice/aslan/core/environment/service"
	"github.com/koderover/zadig/pkg/setting"
	internalhandler "github.com/koderover/zadig/pkg/shared/handler"
	"github.com/koderover/zadig/pkg/shared/kube/resource"
	e "github.com/koderover/zadig/pkg/tool/errors"
	"github.com/koderover/zadig/pkg/tool/log"
	"github.com/koderover/zadig/pkg/types/permission"
)

type UpdateEnvs struct {
	EnvNames   []string `json:"env_names"`
	UpdateType string   `json:"update_type"`
}

type ChartInfoArgs struct {
	ChartInfos []*template.RenderChart `json:"chart_infos"`
}

type NamespaceResource struct {
	Services  []*commonservice.ServiceResp `json:"services"`
	Ingresses []resource.Ingress           `json:"ingresses"`
}

// ListProducts list all product information
func ListProducts(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	ctx.Resp, ctx.Err = service.ListProducts(c.Query("productName"), c.Query("envType"), ctx.User.Name, ctx.User.ID, ctx.User.IsSuperUser, ctx.Logger)
}

func AutoCreateProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	ctx.Resp = service.AutoCreateProduct(c.Param("productName"), c.Query("envType"), ctx.RequestID, ctx.Logger)
}

func AutoUpdateProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(UpdateEnvs)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("AutoUpdateProduct c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("AutoUpdateProduct json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "自动更新", "集成环境", strings.Join(args.EnvNames, ","), fmt.Sprintf("%s,%s", permission.TestEnvManageUUID, permission.ProdEnvManageUUID), string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Logger.Error(err)
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	force, _ := strconv.ParseBool(c.Query("force"))
	ctx.Resp, ctx.Err = service.AutoUpdateProduct(args.EnvNames, c.Param("productName"), ctx.User.ID, ctx.User.IsSuperUser, ctx.RequestID, force, ctx.Logger)
}

func CreateProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(commonmodels.Product)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("CreateProduct c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("CreateProduct json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, args.ProductName, "新增", "集成环境", args.EnvName, fmt.Sprintf("%s,%s", permission.TestEnvCreateUUID, permission.ProdEnvCreateUUID), string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	if args.EnvName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can not be null!")
		return
	}

	//
	args.UpdateBy = ctx.Username
	ctx.Err = service.CreateProduct(
		ctx.Username, ctx.RequestID, args, ctx.Logger,
	)
}

func UpdateProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	args := new(commonmodels.Product)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("UpdateProduct c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("UpdateProduct json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, productName, "更新", "集成环境", envName, fmt.Sprintf("%s,%s", permission.TestEnvManageUUID, permission.ProdEnvManageUUID), string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	force, _ := strconv.ParseBool(c.Query("force"))
	// update product asynchronously
	ctx.Err = service.UpdateProductV2(envName, productName, ctx.Username, ctx.RequestID, force, args.Vars, ctx.Logger)
	if ctx.Err != nil {
		ctx.Logger.Errorf("failed to update product %s %s: %v", envName, productName, ctx.Err)
	}
}

func UpdateProductRecycleDay(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")
	recycleDayStr := c.Query("recycleDay")

	internalhandler.InsertOperationLog(c, ctx.Username, productName, "更新", "集成环境-环境回收", envName, permission.TestEnvManageUUID, "", ctx.Logger)

	var (
		recycleDay int
		err        error
	)
	if recycleDayStr == "" || envName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName or recycleDay不能为空")
		return
	}
	recycleDay, err = strconv.Atoi(recycleDayStr)
	if err != nil || recycleDay < 0 {
		ctx.Err = e.ErrInvalidParam.AddDesc("recycleDay必须是正整数")
		return
	}

	ctx.Err = service.UpdateProductRecycleDay(envName, productName, recycleDay)
}

func UpdateHelmProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	updateType := c.Query("updateType")
	productName := c.Param("productName")

	ctx.Err = service.UpdateHelmProduct(productName, envName, updateType, ctx.Username, ctx.RequestID, ctx.Logger)
	if ctx.Err != nil {
		ctx.Logger.Errorf("failed to update product %s %s: %v", envName, productName, ctx.Err)
	}
}

func UpdateHelmProductVariable(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	args := new(ChartInfoArgs)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("UpdateHelmProductVariable c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("UpdateHelmProductVariable json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "更新", "helm集成环境变量", "", permission.TestEnvManageUUID, string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Logger.Error(err)
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	ctx.Err = service.UpdateHelmProductVariable(productName, envName, ctx.Username, ctx.RequestID, args.ChartInfos, ctx.Logger)
	if ctx.Err != nil {
		ctx.Logger.Errorf("failed to update product Variable %s %s: %v", envName, productName, ctx.Err)
	}
}

func UpdateMultiHelmProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	args := new(UpdateEnvs)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("UpdateMultiHelmProduct c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("UpdateMultiHelmProduct json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "更新helm环境", "集成环境", strings.Join(args.EnvNames, ","), permission.TestEnvManageUUID, string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if err := c.BindJSON(args); err != nil {
		ctx.Logger.Error(err)
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	ctx.Resp = service.UpdateMultiHelmProduct(args.EnvNames, args.UpdateType, c.Param("productName"), ctx.User.ID, ctx.User.IsSuperUser, ctx.RequestID, ctx.Logger)
}

func GetProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	ctx.Resp, ctx.Err = service.GetProduct(ctx.Username, envName, productName, ctx.Logger)
}

func GetProductInfo(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	ctx.Resp, ctx.Err = service.GetProductInfo(ctx.Username, envName, productName, ctx.Logger)
}

func GetProductIngress(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	productName := c.Param("productName")
	ctx.Resp, ctx.Err = service.GetProductIngress(ctx.Username, productName, ctx.Logger)
}

func ListRenderCharts(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	ctx.Resp, ctx.Err = service.ListRenderCharts(productName, envName, ctx.Logger)
	if ctx.Err != nil {
		ctx.Logger.Errorf("failed to listRenderCharts %s %s: %v", envName, productName, ctx.Err)
	}
}

func GetHelmChartVersions(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	ctx.Resp, ctx.Err = service.GetHelmChartVersions(productName, envName, ctx.Logger)
	if ctx.Err != nil {
		ctx.Logger.Errorf("failed to get helmVersions %s %s: %v", envName, productName, ctx.Err)
	}
}

func DeleteProduct(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	envName := c.Query("envName")

	internalhandler.InsertOperationLog(c, ctx.Username, c.Param("productName"), "删除", "集成环境", envName, fmt.Sprintf("%s,%s", permission.TestEnvDeleteUUID, permission.ProdEnvDeleteUUID), "", ctx.Logger)
	ctx.Err = commonservice.DeleteProduct(ctx.Username, envName, c.Param("productName"), ctx.RequestID, ctx.Logger)
}

func EnvShare(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()
	productName := c.Param("productName")

	args := new(service.ProductParams)
	data, err := c.GetRawData()
	if err != nil {
		log.Errorf("CreateProduct c.GetRawData() err : %v", err)
	}
	if err = json.Unmarshal(data, args); err != nil {
		log.Errorf("CreateProduct json.Unmarshal err : %v", err)
	}
	internalhandler.InsertOperationLog(c, ctx.Username, productName, "更新", "集成环境-环境授权", args.EnvName, permission.TestEnvShareUUID, string(data), ctx.Logger)
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	if productName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("productName can't be empty!")
		return
	}

	if err := c.BindJSON(args); err != nil {
		ctx.Err = e.ErrInvalidParam.AddDesc(err.Error())
		return
	}

	if args.EnvName == "" {
		ctx.Err = e.ErrInvalidParam.AddDesc("envName can't be empty!")
		return
	}

	ctx.Err = service.UpdateProductPublic(productName, args, ctx.Logger)
}

func ListGroups(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	productName := c.Param("productName")
	envName := c.Query("envName")
	serviceName := c.Query("serviceName")
	perPageStr := c.Query("perPage")
	perPage := 0
	if perPageStr == "" {
		perPage = setting.PerPage
	} else {
		perPage, _ = strconv.Atoi(perPageStr)
	}

	pageStr := c.Query("page")
	page := 0
	if pageStr == "" {
		page = 1
	} else {
		page, _ = strconv.Atoi(pageStr)
	}

	count := 0
	ctx.Resp, count, ctx.Err = service.ListGroups(serviceName, envName, productName, perPage, page, ctx.Logger)
	c.Writer.Header().Set("X-Total", strconv.Itoa(count))
}

func ListGroupsBySource(c *gin.Context) {
	ctx := internalhandler.NewContext(c)
	defer func() { internalhandler.JSONResponse(c, ctx) }()

	envName := c.Query("envName")
	productName := c.Param("productName")

	services, ingresses, err := commonservice.ListGroupsBySource(envName, productName, ctx.Logger)
	ctx.Resp = &NamespaceResource{
		Services:  services,
		Ingresses: ingresses,
	}
	ctx.Err = err
}

func ListProductsSSE(c *gin.Context) {
	ctx := internalhandler.NewContext(c)

	internalhandler.Stream(c, func(ctx1 context.Context, msgChan chan interface{}) {
		startTime := time.Now()
		wait.NonSlidingUntilWithContext(ctx1, func(_ context.Context) {
			res, err := service.ListProducts(c.Query("productName"), c.DefaultQuery("envType", setting.TestENV), ctx.User.Name, ctx.User.ID, ctx.User.IsSuperUser, ctx.Logger)
			if err != nil {
				ctx.Logger.Errorf("[%s] ListProductsSSE error: %v", ctx.Username, err)
			}

			msgChan <- res

			if time.Since(startTime).Minutes() == float64(60) {
				ctx.Logger.Warnf("[%s] Query ListProductsSSE API over 60 minutes", ctx.Username)
			}
		}, time.Second)
	}, ctx.Logger)
}
