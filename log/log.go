package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/micro/go-micro/v2/config"
	cron "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
)

// Option Log初始化参数
type Option struct {
	Dir          string `json:"filedir"`
	Level        uint32 `json:"level"`
	MaxDays      int64  `json:"maxdays"`
	ReportCaller bool   `json:"report_caller"`
}

var (
	once             sync.Once
	baseLogPath      string                                    // log文件路径
	currLogFile      *os.File                                  // 用于记录前log文件
	currLogFileName  string                                    // 用于记录当前log文件名
	lock             sync.Mutex                                // 切换文件时需要加锁
	option           *Option                                   // log初始化参数
	logTimeFormat    string     = "2006-01-02 15:04:05.000000" //日志时间输入出格式
	changeFileSpec   string     = "0 0 0 */1 * ?"              // 切换日志文件定时任务配置，每天零晨切换
	clearHistorySpec string     = "0 0 1 */1 * ?"              // 清除历史日志定时任务配置，每天1点清理
	lofFileFormat    string     = "2006-01-02"                 // log文件名时间格式，每天一个文件
)

// SendRobotTxtMsg 给钉钉机器人发送消息
func SendRobotTxtMsg(msg string) error {
	sURL := config.Get("robot", "server").String("")
	if sURL == "" || msg == "" {
		return nil
	}
	prefix := config.Get("robot", "prefix").String("")
	switch common.Env {
	case common.EnvDev:
		prefix += "【开发】"
	case common.EnvTest:
		prefix += "【测试】"
	case common.EnvPre:
		prefix += "【预发布】"
	}
	msg = prefix + msg

	data := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]interface{}{
			"content": msg,
		},
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}

	params := &common.HTTPParam{
		URL:     sURL,
		Method:  http.MethodPost,
		Timeout: 2,
		Data:    string(buf),
		Headers: map[string]interface{}{"Content-Type": "application/json;charset=utf-8"},
	}
	resp, code, err := common.HTTP(params)

	if err != nil {
		return err
	}

	if code != 200 {
		return fmt.Errorf("http status code is not 200")
	}

	ret := struct {
		ErrCode int64  `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}{}

	err = json.Unmarshal(resp, &ret)
	if err != nil {
		return nil
	}

	if ret.ErrCode != 0 {
		return fmt.Errorf(ret.ErrMsg)
	}

	return nil
}

// SendRobotTxtMsgHook 发送钉钉消息hook
type SendRobotTxtMsgHook struct {
}

// Fire 发送消息
func (c *SendRobotTxtMsgHook) Fire(entry *logrus.Entry) error {
	msg, _ := entry.String()
	_ = SendRobotTxtMsg(msg)
	return nil
}

// Levels hook等级
func (c *SendRobotTxtMsgHook) Levels() []logrus.Level {
	lever := []logrus.Level{logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel}
	return lever
}

// GetLogOptionFormConfig 初始化log
func GetLogOptionFormConfig() (*Option, error) {
	opt := Option{}
	err := config.Get("log").Scan(&opt)
	if err != nil {
		logrus.Errorf("读取log配置失败")
		return nil, err
	}
	return &opt, nil
}

// InitLogrus 初始化log组件
func InitLogrus(opt *Option) error {
	var err error
	if opt == nil {
		opt, err = GetLogOptionFormConfig()
		if err != nil {
			return err
		}
	}

	_, err = os.Stat(opt.Dir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(opt.Dir, 0755)
			if err != nil {
				logrus.Error("创建日志目录(" + opt.Dir + ")失败。")
				return err
			}
		} else {
			logrus.Error("目录：" + opt.Dir + "，stat失败")
			return fmt.Errorf("目录：" + opt.Dir + "，stat失败")
		}
	}

	// 只做一次
	once.Do(func() {
		lock.Lock()
		option = opt
		logfile := filepath.Base(os.Args[0])
		baseLogPath = opt.Dir + logfile + "."
		lock.Unlock()

		// 设置日志文件
		setLogFile()

		// 定时清理历史日志
		clearHistoryLog()

		// 日志输出文件名、行号、函数名
		logrus.SetReportCaller(opt.ReportCaller)

		//日志格式
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: logTimeFormat,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime: "atime",
			},
			CallerPrettyfier: callerPrettyfier,
		})

		//设置日志等级
		logrus.SetLevel(logrus.Level(opt.Level))

		// 日志添加发钉钉消息hook
		logrus.AddHook(&SendRobotTxtMsgHook{})
	})
	return nil
}

// NewEngineLog 生成引擎专用log
func NewEngineLog(actID, serial, nid string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"act_id":  actID,
		"bserial": serial,
		"bnid":    nid,
	})
}

func setLogFile() {
	var changeFile = func() {
		lock.Lock()
		defer lock.Unlock()

		logFile := baseLogPath + time.Now().Format(lofFileFormat) + ".log"
		if currLogFileName != logFile {
			file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				fmt.Println("open log file error:", err.Error())
				return
			}
			logrus.SetOutput(file)
			if currLogFile != nil {
				currLogFile.Close()
			}
			currLogFileName = logFile
			currLogFile = file
		}
	}

	// 程序启动时设置一次日志文件
	changeFile()

	// 定时切换日志文件
	c := cron.New()
	_, _ = c.AddFunc(changeFileSpec, changeFile)
	c.Start()
}

//  定时清空历史log
func clearHistoryLog() {
	c := cron.New()
	logfile := filepath.Base(os.Args[0])

	var clear = func() {
		lock.Lock()
		defer lock.Unlock()

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
				os.Remove(option.Dir + name) // 删除历史文件
			}
		}
	}

	// 启动时清理一次
	clear()

	_, _ = c.AddFunc(clearHistorySpec, clear)
	c.Start()
}

// callerPrettyfier 格式化log文件名与函数名
func callerPrettyfier(caller *runtime.Frame) (function string, file string) {
	fileName := filepath.Base(caller.File) + ":" + strconv.Itoa(caller.Line)
	funcName := filepath.Base(caller.Function)
	return funcName, fileName
}

// CurrLogFileName 返回当前log文件名
func CurrLogFileName() string {
	lock.Lock()
	defer lock.Unlock()

	return currLogFileName
}
