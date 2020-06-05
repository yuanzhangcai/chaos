// 耗时日志中间简

package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/monitor"
)

// UsedTime 生成耗时日志中间件
func UsedTime() func(c *gin.Context) {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		param := gin.LogFormatterParams{
			Request: c.Request,
			Keys:    c.Keys,
		}

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()
		param.Path = c.Request.URL.Path
		body := c.Request.Form.Encode()
		response, _ := c.Get("response")
		resp, _ := json.Marshal(response)

		// Stop timer
		param.Latency = time.Since(start)

		// 增加监控上报
		monitor.SummaryChaosCostTime(float64(param.Latency))
		monitor.AddURICount(param.Path)

		logrus.Info(fmt.Sprintf("UsedTime: %3d| %13v |%s %-7s %s body[%s] resp[%v] err[%s] size[%d]",
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
			body,
			string(resp),
			param.ErrorMessage,
			param.BodySize,
		))
	}
}
