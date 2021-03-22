package models

import (
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/yuanzhangcai/chaos/common"
	"github.com/yuanzhangcai/config"
)

var (
	dbMap map[string]*gorm.DB = make(map[string]*gorm.DB)
)

// Model 数据库操作组件基类
type Model struct {
	DB  *gorm.DB
	Log *logrus.Entry
}

// SetDB 设置所使用数据库
func (c *Model) SetDB(node string) {
	c.DB = dbMap[node]
}

// Exec 执行sql语句
func (c *Model) Exec(sql string, values ...interface{}) *gorm.DB {
	if c.DB == nil {
		return nil
	}
	return c.DB.Exec(sql, values...)
}

type dbLogger struct {
}

func (c *dbLogger) Print(v ...interface{}) {
	logrus.Info(v...)
}

// ConnectDB 连接db
func ConnectDB(node string) error {
	var err error

	if _, ok := dbMap[node]; ok {
		dbMap[node].Close()
		delete(dbMap, node)
	}

	dbInfo := config.GetString("db", node)
	if dbInfo == "" {
		return fmt.Errorf("没有获取到数据库配置")
	}

	// 初始化连接
	db, err := gorm.Open("mysql", dbInfo)
	if err != nil {
		logrus.Error("数据库初始化失败。错误信息：" + err.Error())
		return err
	}

	// 取消DB复数
	db.SingularTable(true)

	if config.GetBool("db", "write_log") {
		// 设置sql语句输出到日志文件中
		db.LogMode(true)
		logger := &dbLogger{}
		db.SetLogger(logger)
	}

	dbMap[node] = db
	return nil
}

// Init 初始化顾
func Init() error {
	var err error
	list := config.GetStringArray("db", "list")
	if len(list) == 0 {
		// return fmt.Errorf("没有获取到数据库配置")
		return nil
	}

	for _, one := range list {
		err := ConnectDB(one)
		if err != nil {
			logrus.Error("数据库连接失败。", err)
			return err
		}
	}

	return err
}

// GormTime Grom datetime类型
type GormTime struct {
	time.Time
}

// MarshalJSON 数据序列化
func (t GormTime) MarshalJSON() ([]byte, error) {
	formatted := fmt.Sprintf("\"%s\"", t.Format(common.YMDHIS))
	return []byte(formatted), nil
}

// UnmarshalJSON json反序列表
func (t *GormTime) UnmarshalJSON(data []byte) error {
	tm, err := time.Parse(common.YMDHIS, string(data[1:len(data)-1]))
	if err != nil {
		return nil
	}
	*t = GormTime{Time: tm}
	return nil
}

// Value 返回datetime值
func (t GormTime) Value() (driver.Value, error) {
	var zeroTime time.Time
	if t.Time.UnixNano() == zeroTime.UnixNano() {
		return nil, nil
	}
	return t.Time, nil
}

// Scan 设置datetime值
func (t *GormTime) Scan(v interface{}) error {
	value, ok := v.(time.Time)
	if ok {
		*t = GormTime{Time: value}
		return nil
	}
	return fmt.Errorf("can not convert %v to timestamp", v)
}
