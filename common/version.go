package common

import "github.com/yuanzhangcai/config"

const (
	// EnvDev 开发环境
	EnvDev = "dev"
	// EnvTest 测试环境
	EnvTest = "test"
	// EnvPre 测试环境
	EnvPre = "pre"
	// EnvProd 生产环境
	EnvProd = "prod"
)

var (
	// Version 程序版本号
	Version string
	// Env 程序运行环境
	Env string = EnvProd
	// Commit 最后一次提交的id
	Commit string
	// BuildTime 编译时间
	BuildTime string
	// BuildUser 编译人
	BuildUser string
	// GoVersion go编译版本
	GoVersion string
)

// GetVersion 获取版本信息
func GetVersion() map[string]string {
	return map[string]string{
		"app_desc":   config.GetString("common", "app_desc"), //应用描述
		"version":    Version,                                // 程序版本号
		"env":        Env,                                    // 程序运行环境
		"commit":     Commit,                                 // 最后一次提交的id
		"build_time": BuildTime,                              // 编译时间
		"build_user": BuildUser,                              // 编译人
		"go_version": GoVersion,                              // 版本号
	}
}
