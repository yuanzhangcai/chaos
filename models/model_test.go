package models

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/memory"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/chaos/tools"
)

var (
	server   = "10.10.40.49:6379"
	password = "12345678"
	prefix   = ""
)

func init() {
	initConfig()
}

func initConfig() {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/tds/chaos"
	}

	common.Env = common.EnvTest
	common.LoadConfig()

	str := `
	{
		"db" : {
			"list" : ["db1"],
			"db1" : "zacyuan:zacyuan@(10.10.40.49:3306)/tds_user_pre?parseTime=true&loc=Local&charset=utf8",
			"write_log" : true
		}
	}`

	s := memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)

	// 初始化Redis
	_ = tools.InitRedis(server, password, prefix)
}

func TestInit(t *testing.T) {
	str := `
	{
		"db" : {
			"list" : ["bb"],
			"bb" : ""
		}
	}`

	s := memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)
	err := Init()
	assert.NotNil(t, err)

	str = `
	{
		"db" : {
			"list" : ["db1"],
			"db1" : "www"
		}
	}`

	s = memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)
	err = Init()
	assert.NotNil(t, err)

	str = `
	{
		"db" : {
			"list" : []
		}
	}`

	s = memory.NewSource(
		memory.WithJSON([]byte(str)),
	)

	_ = config.Load(s)
	err = Init()
	assert.NotNil(t, err)

	initConfig()

	err = Init()
	assert.Nil(t, err)
}

type TestTime struct {
	Time GormTime `gorm:"Column:dtTime"`
}

func (c TestTime) TableName() string {
	return "tbTestTime"
}

func TestExec(t *testing.T) {
	_ = Init()

	model := &Model{}
	db := model.Exec("delete from tbActQual_19_1 where iUin = ?", 99999)
	assert.Nil(t, db)

	model.SetDB("db1")
	db = model.Exec("delete from tbTestTime;")
	assert.NotNil(t, db)

	params := &TestTime{Time: GormTime{time.Now()}}
	model.DB.Create(params)

	one := TestTime{}
	model.DB.First(&one)

	assert.Equal(t, params.Time.Format(common.YMDHI), one.Time.Format(common.YMDHI))

	buf, err := json.Marshal(params)
	assert.Nil(t, err)

	one2 := TestTime{}
	err = json.Unmarshal(buf, &one2)
	assert.Nil(t, err)
	assert.Equal(t, params.Time.Format(common.YMDHIS), one2.Time.Format(common.YMDHIS))

}
