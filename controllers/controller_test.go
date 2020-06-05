package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/errors"
)

func createController(uri string) (*Controller, *httptest.ResponseRecorder) {
	ctl := &Controller{}
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Request, _ = http.NewRequest("GET", uri, nil)

	ctl.Init(ctx)

	return ctl, w
}

func TestPrepare(t *testing.T) {
	ctl := Controller{}
	if ctl.Prepare() != true {
		t.Fatal("Prepare failed.")
	}
}

func TestControllerInit(t *testing.T) {
	ctl, w := createController("/engine?act_id=19&flow_id=65")

	header := w.Header()
	assert.Equal(t, "19", ctl.Params.Get("act_id"))
	assert.Equal(t, "65", ctl.Params.Get("flow_id"))
	assert.Equal(t, header.Get("Access-Control-Allow-Origin"), "*")
	assert.Equal(t, header.Get("Access-Control-Allow-Methods"), "POST, GET")
}

func TestVersion(t *testing.T) {
	ctl, w := createController("/version")
	ctl.Version()
	assert.Equal(t, `{"data":{"app_desc":"","build_time":"","build_user":"","commit":"","env":"prod","go_version":"","version":""},"msg":"OK","ret":0}`, w.Body.String())
}

func TestOutput(t *testing.T) {
	ctl, w := createController("/version")

	ctl.Output(errors.ErrSystem)
	assert.Equal(t, `{"msg":"系统错误","ret":-9999}`, w.Body.String())
}

func TestOutputJSON(t *testing.T) {
	ctl, w := createController("/version")

	ctl.Result["msg"] = "ok"
	ctl.OutputJSON()
	assert.Equal(t, `{"msg":"ok"}`, w.Body.String())
}
