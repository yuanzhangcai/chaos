# Chaos基础框架

## 使用方法：

```
    go get github.com/yuanzhangcai/chaos
```

复制chaos下的Makefile和config目录到自己工程目录，修改配置文件中的配置参数

新建main.go
```
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/yuanzhangcai/chaos"
	"github.com/yuanzhangcai/chaos/controllers"
	"github.com/yuanzhangcai/chaos/errors"
	"github.com/yuanzhangcai/chaos/services"
)

// DemoCtl demo
type DemoCtl struct {
	controllers.Controller
}

// Demo demo
func (c *DemoCtl) Demo() {
	c.Result["complete"] = 100
	c.Result["canpub"] = 1
	c.Output(errors.OK)
}

func main() {

	SetRouter := func(router *gin.Engine) {
		services.HandleAll(router, "/demo", []string{http.MethodGet, http.MethodPost}, &DemoCtl{}, "Demo")
	}

	// 启动服务
	chaos.Start(SetRouter)
}
```