package chaos

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"

	"github.com/gin-gonic/gin"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micro/go-micro/v2/config"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/log"
	"github.com/yuanzhangcai/chaos/models"
	"github.com/yuanzhangcai/chaos/monitor"
	"github.com/yuanzhangcai/chaos/services"
	"github.com/yuanzhangcai/chaos/tools"
)

func init() {
	// 获取程序运行目录信息
	common.GetRunInfo()

	// 获取当前运行环境
	common.GetEnv()

	// 加载配置文件
	common.LoadConfig()

	// 显示版本信息
	common.ShowInfo()

	// 初始化log
	err := log.InitLogrus(nil)
	if err != nil {
		logrus.Fatal(err)
	}

	// 初始化监控
	monitor.Init()

	// 初始化Redis
	if err = tools.InitRedis(config.Get("redis", "server").String(""),
		config.Get("redis", "password").String(""),
		config.Get("redis", "prefix").String("")); err != nil {
		logrus.Fatal(err)
	}

	// 初始化DB
	if err := models.Init(); err != nil {
		logrus.Fatal(err)
	}
}

// Start 启动服务
func Start(setRouter func(*gin.Engine)) {
	pprof := config.Get("pprof", "server").String("")
	fmt.Println("pprof =", pprof)
	if pprof != "" {
		go func() {
			_ = http.ListenAndServe(pprof, nil) // 非正式环境，开启pprof服务
		}()
	}

	services.Start(setRouter)
}

// Stop 停止服务
func Stop() {
	services.Stop()
}
