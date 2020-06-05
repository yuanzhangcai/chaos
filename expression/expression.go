package expression

import (
	"strconv"
	"time"

	"github.com/Knetic/govaluate"
	"github.com/yuanzhangcai/chaos/common"
)

var functions = map[string]govaluate.ExpressionFunction{
	"strlen": Strlen,
	"intval": Intval,
}

// Strlen 求字符串长度
func Strlen(args ...interface{}) (interface{}, error) {
	length := len(args[0].(string))
	return float64(length), nil
}

// Intval 将字符串转成数字
func Intval(args ...interface{}) (interface{}, error) {
	val := common.ParseInt64(args[0].(string))
	return float64(val), nil
}

// Date 返回当前日期
func Date() string {
	return time.Now().Format(common.YMD)
}

// Time 返回当前时间戳字符符串
func Time() string {
	return strconv.FormatInt(time.Now().Unix(), 10)
}

// Eval 执行表达式
func Eval(exp string) (interface{}, error) {
	expr, err := govaluate.NewEvaluableExpressionWithFunctions(exp, functions)
	if err != nil {
		return nil, err
	}
	return expr.Evaluate(nil)
}
