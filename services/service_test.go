package services

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/controllers"
	"github.com/yuanzhangcai/chaos/monitor"
)

type header struct {
	Key   string
	Value string
}

type prepareCtl struct {
	controllers.Controller
}

func (c *prepareCtl) Prepare() bool {
	c.Ctx.String(200, "OK")
	return false
}

func (c *prepareCtl) Test() {
	c.Ctx.String(200, "Test")
}

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

func performRequest(r http.Handler, method, path string, headers ...header) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	for _, h := range headers {
		req.Header.Add(h.Key, h.Value)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func checkRouters500(t *testing.T, r *gin.Engine, uri string) {
	w := performRequest(r, http.MethodGet, uri)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	w = performRequest(r, http.MethodPost, uri)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func checkExistRouters(t *testing.T, r *gin.Engine, uri string) {
	w := performRequest(r, http.MethodGet, uri)
	buf, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, "404 page not found", string(buf))

	w = performRequest(r, http.MethodPost, uri)
	buf, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEqual(t, "404 page not found", string(buf))
}

func checkRoutersEqual(t *testing.T, r *gin.Engine, uri string, code int, cmpVal string) {
	w := performRequest(r, http.MethodGet, uri)
	buf, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, code, w.Code)
	assert.Equal(t, cmpVal, string(buf))
}

func checkNotExistRouters(t *testing.T, r *gin.Engine, uri string) {
	w := performRequest(r, http.MethodGet, uri)
	buf, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "404 page not found", string(buf))

	w = performRequest(r, http.MethodPost, uri)
	buf, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "404 page not found", string(buf))

	w = performRequest(r, http.MethodPut, uri)
	buf, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "404 page not found", string(buf))
}

func checkNoMethod(t *testing.T, r *gin.Engine, uri string) {
	w := performRequest(r, http.MethodGet, uri)
	buf, _ := ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "NoMethod is not exist.", string(buf))

	w = performRequest(r, http.MethodPost, uri)
	buf, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "NoMethod is not exist.", string(buf))

	w = performRequest(r, http.MethodPut, uri)
	buf, _ = ioutil.ReadAll(w.Body)
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Equal(t, "404 page not found", string(buf))
}

func TestServer(t *testing.T) {
	initConfig()

	r := CreateServer()
	assert.NotEqual(t, r, nil)

	CreateRouters(r)

	checkExistRouters(t, r, "/version")

	checkNotExistRouters(t, r, "/yy/bb")
	checkNotExistRouters(t, r, "/html/yyy.html")

	HandleAll(r, "/router", []string{http.MethodGet, http.MethodPost}, &controllers.Controller{}, "Version")
	checkExistRouters(t, r, "/router")

	group := r.Group("/chaos")
	HandleAll(group, "/version", []string{http.MethodGet, http.MethodPost}, &controllers.Controller{}, "Version")
	checkExistRouters(t, r, "/chaos/version")

	str := "not router"
	HandleAll(str, "/not_router", []string{http.MethodGet, http.MethodPost}, &controllers.Controller{}, "Version")
	checkNotExistRouters(t, r, "/not_router")

	HandleAll(r, "/prepare", []string{http.MethodGet, http.MethodPost}, &prepareCtl{}, "Test")
	checkRoutersEqual(t, r, "/prepare", 200, "OK")

	HandleAll(r, "/panic", []string{http.MethodGet, http.MethodPost}, &struct{}{}, "Version")
	checkRouters500(t, r, "/panic")

	HandleAll(r, "/no_method", []string{http.MethodGet, http.MethodPost}, &controllers.Controller{}, "NoMethod")
	checkNoMethod(t, r, "/no_method")
}

func TestStart(t *testing.T) {
	initConfig()

	go Start(func(router *gin.Engine) {})

	time.Sleep(2 * time.Second)

	result, code, err := common.GetHTTP("http://127.0.0.1:4004/version")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, http.StatusOK, code)
	assert.Equal(t, `{"data":{"app_desc":"chaos","build_time":"","build_user":"","commit":"","env":"test","go_version":"","version":""},"msg":"OK","ret":0}`, string(result))

	Stop()

	time.Sleep(1 * time.Second)

	_, _, err = common.GetHTTP("http://127.0.0.1:4004/version")
	assert.NotEqual(t, nil, err)

	str := `
	{
		"common" : {
			"etcd_addrs" : ["47.99.79.44:2378"]
		}
	}`

	s := memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)

	Start(func(router *gin.Engine) {})
}
