package monitor

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/config"
)

func initConfig() {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}

	common.Env = "test"
	common.LoadConfig()

	str := `
	{
		"common" : {
			"etcd_addrs" : ["47.99.79.44:2379", "47.111.108.59:2379", "47.99.62.229:2379"]
		}
	}`

	_ = config.LoadMemory(str, "json")
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
