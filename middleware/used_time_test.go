package middleware

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/log"
	"github.com/yuanzhangcai/chaos/monitor"
)

func initConfig() {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/tds/chaos"
	}

	common.Env = "test"
	common.LoadConfig()

	str := `
	{
		"common" : {
			"etcd_addrs" : ["47.99.79.44:2379", "47.111.108.59:2379", "47.99.62.229:2379"]
		}
	}`

	s := memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)

	monitor.SetMetrics()
}

func performRequest(r http.Handler, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUsedTime(t *testing.T) {
	initConfig()

	// 初始化log
	opt := log.Option{
		Dir:          common.CurrRunPath + "/logs/",
		MaxDays:      15,
		Level:        4,
		ReportCaller: true,
	}
	_ = log.InitLogrus(&opt)

	r := gin.New()

	// 添加中间件
	var ware []gin.HandlerFunc
	ware = append(ware, UsedTime())
	ware = append(ware, gin.Recovery())
	r.Use(ware...)
	r.GET("/middleware", func(ctx *gin.Context) {
		ctx.String(200, "middleware")
	})

	w := performRequest(r, http.MethodGet, "/middleware")
	buf, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, "404 page not found", string(buf))

	buf, err := ioutil.ReadFile(log.CurrLogFileName())
	assert.Nil(t, err)

	logStr := string(buf)
	assert.Contains(t, logStr, "UsedTime")
}
