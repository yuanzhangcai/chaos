package monitor

import (
	"os"
	"testing"
	"time"

	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
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
}

func init() {
	time.Sleep(1 * time.Second)
	initConfig()
}

func TestInit(t *testing.T) {
	Init()
	assert.NotNil(t, register)
	assert.NotNil(t, srv)
}

func TestStop(t *testing.T) {
	Init()

	Stop()
	assert.Nil(t, register)
}

func TestMetrics(t *testing.T) {
	Init()

	AddActVisitCount("10", "aa")

	SummaryChaosCostTime(234)

	AddURICount("/engine")
}
