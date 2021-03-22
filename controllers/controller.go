package controllers

import (
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/errors"
)

// ControllerInterface Controller接口定义
type ControllerInterface interface {
	Init(*gin.Context)
	Prepare() bool
	// Finish()
}

// Controller 逻辑控制处理器基类组件
type Controller struct {
	Ctx    *gin.Context
	Params *url.Values
	Result map[string]interface{} // 返回给前端的数据
}

// Prepare 在主逻辑处理之前的前置操作
// return true 续继后面的操作
//        false 逻辑处理提前结束
func (c *Controller) Prepare() bool {
	return true
}

// Finish 在主逻辑处理之前的收尾操作
// func (c *Controller) Finish() {
// }

// Init 设置Context
func (c *Controller) Init(ctx *gin.Context) {
	c.Result = make(map[string]interface{})
	c.Ctx = ctx
	err := c.Ctx.Request.ParseForm()
	if err != nil {
		logrus.Panic("parse from failed")
	}

	// 所有输入参数去空格
	// 所有输入参数去空格
	params := url.Values{}
	for k, v := range c.Ctx.Request.Form {
		for _, one := range v {
			params.Add(k, strings.TrimSpace(one))
		}
	}
	c.Params = &params

	// c.Ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET")
}

// Version 返回当前版本信息
func (c *Controller) Version() {
	c.Result["data"] = common.GetVersion()
	c.Output(errors.OK)
}

// Output 输入出json
func (c *Controller) Output(ret *errors.Error) {
	c.Result["ret"] = ret.Code()
	c.Result["msg"] = ret.Msg()
	c.OutputJSON()
}

// OutputJSON 将参数直接输出为json
func (c *Controller) OutputJSON() {
	c.Ctx.JSON(200, c.Result)
	c.Ctx.Set("response", c.Result)
}
