package monitor

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/config"
)

var (
	// Namespace 全局命名空间
	Namespace string
	// Subsystem 子系统
	Subsystem string
)

var (
	// IP 当前机器IP
	IP   string
	once sync.Once
	srv  *http.Server

	// actVisitCount 活动访问量
	actVisitCount *prometheus.CounterVec

	// chaosCostTime chaos总耗时情况统计
	chaosCostTime *prometheus.SummaryVec

	// uriCount 各uri调用资数
	uriCount *prometheus.CounterVec
)

// SetMetrics 设置监控指标
func SetMetrics() {
	once.Do(func() {
		// 获取本机IP
		IP = common.GetIntranetIP()

		env := ""
		if common.Env != common.EnvProd {
			env = "_" + common.Env
		}

		// actVisitCount 活动访问量
		actVisitCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "act_visit_count" + env,
				Help:      "act visit count.",
			},
			[]string{"ip", "act_id", "act_name"},
		)

		chaosCostTime = prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  Namespace,
				Subsystem:  Subsystem,
				Name:       "cost_time_seconds" + env,
				Help:       "chaos const time.",
				Objectives: map[float64]float64{0.5: 0.05, 0.7: 0.03, 0.8: 0.02, 0.9: 0.01, 0.99: 0.001},
			},
			[]string{"ip"},
		)

		// uriCount 总访问量
		uriCount = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: Subsystem,
				Name:      "uri_count" + env,
				Help:      "uri count",
			},
			[]string{"ip", "uri"},
		)

		// 注册监控指标
		prometheus.MustRegister(
			actVisitCount,
			chaosCostTime,
			uriCount,
		)
	})
}

// Init 初始化prometheus监控
func Init() {
	Namespace = config.GetString("monitor", "namespace")
	Subsystem = config.GetString("monitor", "subsystem")

	if srv != nil {
		return
	}

	// 设置监控指标
	SetMetrics()

	addr := config.GetString("monitor", "server")
	if addr != "" { // 开启prometheus监控
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		srv = &http.Server{}
		srv.Addr = addr
		srv.Handler = mux

		go func() {
			// http.Handle("/metrics", promhttp.Handler())
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logrus.Fatalf("listen: %s\n", err)
			}
		}()
	}
}

// Stop 停止监控上报
func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	logrus.Info("Monitor Shutdown Server ...")
	if srv != nil {
		if err := srv.Shutdown(ctx); err != nil {
			logrus.Error("Server Shutdown:", err)
		}
	}
	logrus.Info("Monitor Server exiting")
}

// AddActVisitCount 总访问量加1
func AddActVisitCount(actID, actName string) {
	actVisitCount.WithLabelValues(IP, actID, actName).Inc()
}

// SummaryChaosCostTime 统计接口调用情况
func SummaryChaosCostTime(v float64) {
	chaosCostTime.WithLabelValues(IP).Observe(v)
}

// AddURICount uri访问量加1
func AddURICount(uri string) {
	uriCount.WithLabelValues(IP, uri).Inc()
}
