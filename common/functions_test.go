package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yuanzhangcai/config"
)

// 自己测试时需要设置环境变量CI_PROJECT_DIR=代码路径，如export CI_PROJECT_DIR=/Users/zacyuan/MyWork/chaos

func TestInit(t *testing.T) {
	assert.NotNil(t, localCache)
	assert.NotNil(t, defaultTransport)
	assert.Empty(t, CurrRunPath)
	assert.Empty(t, CurrRunFileName)
	assert.Equal(t, "2006", Y)
	assert.Equal(t, "2006-01", YM)
	assert.Equal(t, "2006-01-02", YMD)
	assert.Equal(t, "20060102", YMD2)
	assert.Equal(t, "2006-01-02 15", YMDH)
	assert.Equal(t, "2006-01-02 15:04", YMDHI)
	assert.Equal(t, "200601021504", YMDHI2)
	assert.Equal(t, "2006-01-02 15:04:05", YMDHIS)
	assert.Equal(t, "20060102150405", YMDHIS2)
	assert.Equal(t, "15:04", HI)
	assert.Equal(t, "1504", HI2)
}

func TestGetRunInfo(t *testing.T) {
	GetRunInfo()
	assert.NotEmpty(t, CurrRunPath)
	assert.NotEmpty(t, CurrRunFileName)
}

func TestTimeToStr(t *testing.T) {

	t.Run("int", func(t *testing.T) {
		val := 1588221555
		str := TimeToStr(YMDHIS, val)

		assert.Equal(t, "2020-04-30 12:39:15", str)
	})

	t.Run("int64", func(t *testing.T) {
		var val int64 = 1588221555
		str := TimeToStr(YMDHIS, val)
		assert.Equal(t, "2020-04-30 12:39:15", str)
	})

	t.Run("string", func(t *testing.T) {
		val := "1588221555"
		str := TimeToStr(YMDHIS, val)
		assert.Equal(t, "2020-04-30 12:39:15", str)
	})

	t.Run("bool", func(t *testing.T) {
		val := false
		str := TimeToStr(YMDHIS, val)
		assert.Equal(t, "", str)
	})
}

func TestStrToTime(t *testing.T) {
	str := "2020-04-30 12:39:15"
	val := StrToTime(YMDHIS, str)
	assert.Equal(t, int64(1588221555), val)
}

func TestParseInt64(t *testing.T) {
	assert.Equal(t, int64(12345435), ParseInt64("12345435"))
}

func TestParseUint64(t *testing.T) {
	assert.Equal(t, uint64(12345435), ParseUint64("12345435"))
}

func TestMd5Str(t *testing.T) {
	str := "zacyuan changcai yuanzhangcai"
	md5 := "c0104977798896ba979d7965ecde226b"
	assert.Equal(t, md5, Md5Str(str))
}

func TestMd5Byte(t *testing.T) {
	str := "zacyuan changcai yuanzhangcai"
	md5 := "c0104977798896ba979d7965ecde226b"
	value := fmt.Sprintf("%x", Md5Byte(str))
	assert.Equal(t, md5, value)
}

func TestDecimal(t *testing.T) {
	var src float64 = 1.23456
	value := Decimal(src, 2)
	assert.Equal(t, 1.23, value)

	value = Decimal(src, 3)
	assert.Equal(t, 1.235, value)
}

func TestGetRandomString(t *testing.T) {
	for i := 0; i < 10; i++ {
		one := GetRandomString(5)
		two := GetRandomString(5)
		assert.NotEqual(t, one, two)
	}
}

func TestGetFileNameWithoutSuffix(t *testing.T) {
	assert.Equal(t, "config", GetFileNameWithoutSuffix("/data/tds/outer/chaos/config.toml"))
}

type RuntimeError struct{}

func (c *RuntimeError) Error() string {
	return "RuntimeError"
}

func (c *RuntimeError) RuntimeError() {}

func TestToString(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		assert.Equal(t, "changcai", ToString("changcai"))
	})

	t.Run("int", func(t *testing.T) {
		v := 435
		assert.Equal(t, "435", ToString(v))
	})

	t.Run("uint", func(t *testing.T) {
		var v uint = 43545
		assert.Equal(t, "43545", ToString(v))
	})

	t.Run("uint8", func(t *testing.T) {
		var v uint8 = 123
		assert.Equal(t, "123", ToString(v))
	})

	t.Run("int64", func(t *testing.T) {
		var v int64 = 4354546
		assert.Equal(t, "4354546", ToString(v))
	})

	t.Run("uint64", func(t *testing.T) {
		var v uint64 = 43545467
		assert.Equal(t, "43545467", ToString(v))
	})

	t.Run("float64", func(t *testing.T) {
		var v float64 = 43545467.56
		assert.Equal(t, "43545467.56", ToString(v))
	})

	t.Run("bool", func(t *testing.T) {
		assert.Equal(t, "true", ToString(true))
		assert.Equal(t, "false", ToString(false))
	})

	t.Run("runtime.Error", func(t *testing.T) {
		var v runtime.Error = &RuntimeError{}
		assert.Equal(t, "RuntimeError", ToString(v))
	})

	t.Run("error", func(t *testing.T) {
		err := errors.New("changcai")
		assert.Equal(t, "changcai", ToString(err))
	})

	t.Run("other", func(t *testing.T) {
		v := RuntimeError{}
		assert.Empty(t, ToString(v))
	})
}

func TestCreateTransport(t *testing.T) {
	params := HTTPParam{}
	transport := createTransport(&params)
	assert.NotNil(t, transport)

	newTransport := createTransport(&params)
	assert.Equal(t, transport, newTransport)

	params.UseShort = true
	newTransport = createTransport(&params)
	assert.NotEqual(t, transport, newTransport)
}

func TestGetHTTP(t *testing.T) {
	t.Run("HTTP1", func(t *testing.T) {
		result, code, err := GetHTTP("https://tds-test.papegames.com/amsIndex.php?c=clientConfig&a=HttpTest")
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, "HttpTest", string(result))
	})

	t.Run("HTTP2", func(t *testing.T) {
		params := &HTTPParam{
			URL:  "https://tds-test.papegames.com/amsIndex.php",
			Data: "c=clientConfig&a=HttpTest",
			Headers: map[string]interface{}{
				"aa": "11",
				"bb": "22",
			},
			Cookies: map[string]interface{}{
				"cc": "33",
				"dd": "44",
			},
		}
		result, code, err := HTTP(params)
		assert.Nil(t, err)
		assert.Equal(t, 200, code)
		assert.Equal(t, "HttpTest11223344", string(result))
	})

	t.Run("HTTP3", func(t *testing.T) {
		_, code, err := GetHTTP("https://tds-test.papegames.com/amsIndex1.php?c=clientConfig&a=HttpTest")
		assert.Nil(t, err)
		assert.Equal(t, 404, code)
	})

	t.Run("HTTP4", func(t *testing.T) {
		_, _, err := GetHTTP("https://tds-test1.papegames.com/amsIndex.php?c=clientConfig&a=HttpTest")
		assert.NotNil(t, err)
	})
}

func TestGeneratePigeonSig(t *testing.T) {
	secret := "changcai"
	currTime, sig := GeneratePigeonSig(secret)

	sig2 := Md5Str(secret + currTime)
	assert.Equal(t, sig, sig2)
}

func TestCheckFlowAccessLimitLocal(t *testing.T) {
	nid := "1707357"
	actID := "1"
	flowID := "1"
	var second int64 = 0
	var limit int64 = 0

	ret := CheckFlowAccessLimitLocal(nid, actID, flowID, second, limit)
	assert.True(t, ret)

	second = 1
	limit = 3

	ret = CheckFlowAccessLimitLocal(nid, actID, flowID, second, limit)
	assert.True(t, ret)

	ret = CheckFlowAccessLimitLocal(nid, actID, flowID, second, limit)
	assert.True(t, ret)

	ret = CheckFlowAccessLimitLocal(nid, actID, flowID, second, limit)
	assert.True(t, ret)

	ret = CheckFlowAccessLimitLocal(nid, actID, flowID, second, limit)
	assert.False(t, ret)
}

func TestGetIntranetIP(t *testing.T) {
	ip := GetIntranetIP()
	assert.NotEmpty(t, ip)
}

func TestGetEnv(t *testing.T) {
	CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if CurrRunPath == "" {
		CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}

	_ = ioutil.WriteFile(CurrRunPath+"/config/env", []byte("pre"), 0755)

	GetEnv()
	assert.Equal(t, "pre", Env)

	_ = os.Remove(CurrRunPath + "/config/env")
}

func TestLoadConfig(t *testing.T) {
	CurrRunPath = os.Getenv("CI_PROJECT_DIR")
	if CurrRunPath == "" {
		CurrRunPath = "/Users/zacyuan/MyWork/chaos"
	}

	Env = "test"
	LoadConfig()

	Env = "bb"
	LoadConfig()

	tmp := config.GetString("common", "server_name")
	assert.Equal(t, "chaos.zacyuan.com", tmp)
}

func TestShowInfo(t *testing.T) {
	ShowInfo()
}
