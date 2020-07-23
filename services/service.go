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
	"github.com/yuanzhangcai/chaos/tools"
	"github.com/yuanzhangcai/config"
)

var register *tools.ServicesRegister
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

// micro 问题较多，一个进程同时起启两个服务会出现数据竞态问题，使用etcd做服务发现时，也会偶现竞态问题。所以弃用。直接规换成：gin+服务注册。
// // startMicro 启动web服务
// func startMicro(router *gin.Engine, srv *http.Server, ctx *context.Context) {
// 	serverName := config.Get("micro", "server_name").String("chaos.papegames.com") // 微务服名称
// 	if common.Env != common.EnvProd {
// 		serverName += "." + common.Env // 如果当前环境不是正式环境，服务名称添加环境后缀
// 	}

// 	var opt = []web.Option{
// 		web.Name(serverName),    // 设置服务名称
// 		web.HandleSignal(false), // 关闭micro信号处理功能
// 		web.Context(*ctx),       // 设置context
// 		web.Server(srv),         // 设置http server
// 		web.Handler(router),     // 注册Handler事件
// 		web.RegisterInterval(time.Duration(config.Get("micro", "register_interval").Int(15)) * time.Second), // 服务注册间隔时间
// 		web.RegisterTTL(time.Duration(config.Get("micro", "register_ttl").Int(30)) * time.Second),           // 服务失效时间
// 	}

// 	if config.Get("common", "address").String("") != "" {
// 		opt = append(opt, web.Address(config.Get("common", "address").String("")))
// 	}

// 	// 用etcdv3做服务注册与发现
// 	register := etcd.NewRegistry(func(op *registry.Options) {
// 		etcdAdds := config.Get("micro", "etcd_addrs").StringSlice([]string{})
// 		op.Addrs = etcdAdds
// 	})
// 	opt = append(opt, web.Registry(register))

// 	// 创建微服务
// 	service := web.NewService(opt...)

// 	// 服务初始化
// 	err := service.Init()
// 	if err != nil {
// 		logrus.Fatal("服务初始化失败。")
// 	}

// 	go func() {
// 		// Run server
// 		if err := service.Run(); err != nil {
// 			logrus.Fatal(err)
// 		}
// 	}()
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
		register = tools.NewServicesRegister(&tools.RegisterOptions{
			ServerName:    serverName,
			EtcdAddress:   config.GetStringArray("common", "etcd_addrs"),
			ServerAddress: srv.Addr,
			Interval:      time.Duration(config.GetInt("common", "register_interval")) * time.Second,
			TTL:           time.Duration(config.GetInt("common", "register_ttl")) * time.Second,
		})

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
