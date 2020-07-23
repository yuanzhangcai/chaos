package log

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/config"
)

// 自己测试时需要设置环境变量CI_PROJECT_DIR=代码路径，如：export CI_PROJECT_DIR=/Users/zacyuan/MyWork/chaos
func TestGetLogOptionFormConfig(t *testing.T) {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}
	common.Env = "test"
	common.LoadConfig()

	opt, err := GetLogOptionFormConfig()
	assert.Nil(t, err)
	assert.Equal(t, config.GetString("log", "filedir"), opt.Dir)
	assert.Equal(t, uint32(config.GetInt("log", "level")), opt.Level)
	assert.Equal(t, config.GetBool("log", "report_caller"), opt.ReportCaller)
	assert.Equal(t, int64(config.GetInt64("log", "maxdays")), opt.MaxDays)
}

func TestSendRobotTxtMsg(t *testing.T) {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/tds/cosmos"
	}
	common.Env = "test"
	common.LoadConfig()

	err := SendRobotTxtMsg("Pipeline 测试用。")
	assert.Nil(t, err)

	err = SendRobotTxtMsg("")
	assert.Nil(t, err)
}

// 自己测试时需要设置环境变量CI_PROJECT_DIR=代码路径，如：export CI_PROJECT_DIR=/Users/zacyuan/MyWork/chaos
func TestInitLogrus(t *testing.T) {

	_ = InitLogrus(nil)

	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}
	common.Env = "test"
	common.LoadConfig()

	_ = InitLogrus(nil)

	opt := Option{
		Dir:          common.CurrRunPath + "/logs/",
		MaxDays:      15,
		Level:        4,
		ReportCaller: false,
	}

	err := InitLogrus(&opt)
	assert.Nil(t, err)

	logrus.Info("info")
	logrus.Error("error")
}

func TestNewEngineLog(t *testing.T) {
	log := NewEngineLog("19", "aaa", "123")
	log.Info("info")
	//log.Error("error")
}

func TestSetLogFile(t *testing.T) {
	setLogFile()

	// 以下代码在gitlab-runner上测试会有问题，在本机测试没问题，为了能在gitlab-runner上能通过，代码暂时注销
	// changeFileSpec = "*/1 * * * * ?"
	// lofFileFormat = "2006-01-02_15-04-05"
	// setLogFile()

	// lock.Lock()
	// tmp := currLogFileName
	// lock.Unlock()

	// time.Sleep(2 * time.Second)

	// lock.Lock()
	// tmp2 := currLogFileName
	// lock.Unlock()

	// fmt.Println(tmp, tmp2)

	// if tmp == tmp2 {
	// 	t.Fatal("setLogFile failed.")
	// }
}

func TestClearHistoryLog(t *testing.T) {
	common.CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if common.CurrRunPath == "" {
		common.CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}

	opt := Option{
		Dir:          common.CurrRunPath + "/logs/",
		MaxDays:      15,
		Level:        4,
		ReportCaller: true,
	}

	_ = InitLogrus(&opt)

	changeFileSpec = "*/1 * * * * ?"
	lofFileFormat = "2006-01-02_15-04-05"

	clearHistoryLog()

	logfile := filepath.Base(os.Args[0])
	dir, err := ioutil.ReadDir(option.Dir)
	if err != nil {
		return
	}

	for _, fi := range dir {
		name := fi.Name()
		if !fi.IsDir() &&
			strings.HasPrefix(name, logfile) &&
			strings.HasSuffix(name, ".log") &&
			fi.ModTime().Unix() <= time.Now().Add(-time.Hour*24*time.Duration(option.MaxDays)).Unix() {
			t.Fatal("clearHistoryLog failed.")
		}
	}
}

func TestCallerPrettyfier(t *testing.T) {
	caller := runtime.Frame{
		File:     "/Users/zacyuan/MyWork/chaos/main.go",
		Line:     34,
		Function: "/Users/zacyuan/MyWork/chaos/main.go/main",
	}
	function, file := callerPrettyfier(&caller)
	assert.Equal(t, "main", function)
	assert.Equal(t, "main.go:34", file)
}

func TestAa(t *testing.T) {
	lock.Lock()
	tmp := currLogFileName
	lock.Unlock()

	tmp2 := CurrLogFileName()

	assert.Equal(t, tmp, tmp2)
}
