package services

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/controllers"
	"github.com/yuanzhangcai/chaos/middleware"
	"github.com/yuanzhangcai/chaos/monitor"
	"github.com/yuanzhangcai/config"
	"github.com/yuanzhangcai/srsd/registry"
	"github.com/yuanzhangcai/srsd/service"
)

var register *registry.Registry
var quit chan os.Signal

// CreateServer 创建路由
func CreateServer() *gin.Engine {
	if common.Env == common.EnvProd {
		// 正式环境时，将gin的模式，设置成ReleaseMode
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 添加中间件
	var ware []gin.HandlerFunc
	if config.GetBool("common", "used_time") {
		ware = append(ware, middleware.UsedTime())
	}
	ware = append(ware, gin.Recovery())
	router.Use(ware...)
	return router
}

// CreateRouters 创建路由规则
func CreateRouters(router *gin.Engine) {

	// // 静态文件路由
	// router.GET("/html/*filepath", NewBindataHandler("html"))
	// router.POST("/html/*filepath", NewBindataHandler("html"))

	// 设置获取版本信息接口路由
	HandleAll(router, "/version", []string{http.MethodGet, http.MethodPost}, &controllers.Controller{}, "Version")
}

type handleFun func(httpMethod, relativePath string, handlers ...gin.HandlerFunc) gin.IRoutes

// HandleAll 批量设置路由
func HandleAll(r interface{}, relativePath string, httpMethods []string, ctl interface{}, method string) {
	var handle handleFun
	switch h := r.(type) {
	case *gin.Engine:
		handle = h.Handle
	case *gin.RouterGroup:
		handle = h.Handle
	default:
		return
	}

	for _, httpMethod := range httpMethods {
		handle(httpMethod, relativePath, HandleMain(ctl, method))
	}
}

// HandleMain 主要处理逻构造方法
func HandleMain(ctl interface{}, method string) func(*gin.Context) {
	return func(ctx *gin.Context) {
		// 初始一个新的处理器
		temp := initialize(ctl)
		if temp == nil {
			panic("controller is not ControllerInterface")
		}

		// 设置请求上下文
		temp.Init(ctx)

		// 逻辑提前结束
		if !temp.Prepare() {
			return
		}

		// 主处理逻辑
		value := reflect.ValueOf(temp)
		main := value.MethodByName(method)
		if main.IsValid() {
			main.Call(nil)
		} else {
			ctx.String(404, method+" is not exist.")
			return
		}

		// 最后收尾工作
		// temp.Finish()
	}
}

func initialize(c interface{}) controllers.ControllerInterface {

	reflectVal := reflect.ValueOf(c)
	t := reflect.Indirect(reflectVal).Type()

	vc := reflect.New(t)
	execController, ok := vc.Interface().(controllers.ControllerInterface)
	if !ok {
		return nil
	}

	// 对象复制
	// elemVal := reflect.ValueOf(c).Elem()
	// elemType := reflect.TypeOf(c).Elem()
	// execElem := reflect.ValueOf(execController).Elem()

	// numOfFields := elemVal.NumField()
	// for i := 0; i < numOfFields; i++ {
	// 	fieldType := elemType.Field(i)
	// 	elemField := execElem.FieldByName(fieldType.Name)
	// 	if elemField.CanSet() {
	// 		fieldVal := elemVal.Field(i)
	// 		elemField.Set(fieldVal)
	// 	}
	// }

	return execController
}

// // NewBindataHandler 生成bindata handler
// func NewBindataHandler(baseDir string) func(*gin.Context) {
// 	return func(ctx *gin.Context) {
// 		p := ctx.Param("filepath")
// 		p = baseDir + "/" + p
// 		p = filepath.Dir(p) + "/" + filepath.Base(p)
// 		data, err := bindata.Asset(p)
// 		if err != nil {
// 			ctx.String(404, "404 page not found")
// 			return
// 		}

// 		_, _ = ctx.Writer.Write(data)
// 	}
// }

// StartGin 开启gin服务
func StartGin(router *gin.Engine, srv *http.Server) {
	serverName := config.GetString("common", "server_name") // 微务服名称
	if common.Env != common.EnvProd {
		serverName += "." + common.Env // 如果当前环境不是正式环境，服务名称添加环境后缀
	}

	srv.Addr = config.GetString("common", "address")
	srv.Handler = router

	go func() {
		// 服务连接
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("listen: %s\n", err)
			close(quit) // 关闭服务
		}
	}()

	etcdAddrs := config.GetStringArray("common", "etcd_addrs")
	if len(etcdAddrs) > 0 { // 配有etcd地址，则开启服务注册功能
		info := service.NewService()
		info.Name = serverName
		info.Host = srv.Addr
		info.Metrics = config.GetString("monitor", "server")
		info.PProf = config.GetString("pprof", "server")
		register = registry.NewRegistry(info,
			registry.Addresses(config.GetStringArray("common", "etcd_addrs")),
			registry.TTL(time.Duration(time.Duration(config.GetInt("common", "register_ttl")))*time.Second))
		err := register.Start()
		if err != nil {
			fmt.Println("服务注册失败:", err)
			logrus.Error("服务注册失败:", err)
			close(quit) // 关闭服务
		}
	}
}

// StartServer 启动服务
func StartServer(router *gin.Engine) {
	quit = make(chan os.Signal, 1)
	srv := &http.Server{}
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	microCtx, microcancel := context.WithCancel(context.Background())
	defer microcancel()

	// logrus.Info("mode = " + common.Mode)
	// if common.Mode == "micro" {
	// 	// 以微服务型式启动
	// 	startMicro(router, srv, &microCtx)
	// } else {
	// 	// 以传统web服务启动
	// 	startGin(router, srv)
	// }

	// 以传统web服务启动
	StartGin(router, srv)

	<-quit // 等待退出信号

	if register != nil {
		// 停止服务注册
		_ = register.Stop()
	}

	// 关闭服务，设置超时时间为5秒
	ctx, cancel := context.WithTimeout(microCtx, 5*time.Second)
	defer cancel()
	logrus.Info("Shutdown Server ...")
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Error("Server Shutdown:", err)
	}

	// 停止监控
	monitor.Stop()

	quit = nil

	logrus.Info("Server exiting")
	fmt.Println("Server exiting")
}

// Start 开启服务
func Start(setRouter func(router *gin.Engine)) {
	// 创建服务
	router := CreateServer()

	// 创建路由规则
	CreateRouters(router)

	setRouter(router)

	// 开启服务
	StartServer(router)
}

// Stop 停止服务
func Stop() {
	if quit != nil {
		close(quit)
	}
}
