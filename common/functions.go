package common

import (
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/micro/go-micro/v2/config"
	"github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

var (
	localCache       *cache.Cache    // 用于本地频率控制缓存
	defaultTransport *http.Transport // 全局变变更，用于保存长链接缓存。
)

func init() {
	localCache = cache.New(1*time.Minute, 1*time.Minute)
	defaultTransport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
}

// GetRunInfo 获取程序运行信息
func GetRunInfo() {
	ex, err := os.Executable()
	if err != nil {
		os.Exit(-1)
	}

	// 获取当前程序运行文件目录与文件名
	CurrRunPath = filepath.Dir(ex)
	CurrRunFileName = GetFileNameWithoutSuffix(ex)
}

// TimeToStr 时间戳转日期
func TimeToStr(fmt string, value interface{}) string {
	str := ""
	var sec int64
	switch t := value.(type) {
	case int:
		sec = int64(t)
	case int64:
		sec = t
	case string:
		sec, _ = strconv.ParseInt(t, 10, 64)
	default:
		return ""
	}

	str = time.Unix(sec, 0).Format(fmt)
	return str
}

// StrToTime 日期转时间戳
func StrToTime(fmt string, value string) int64 {
	tm, _ := time.ParseInLocation(fmt, value, time.Local)
	return tm.Unix()
}

// ParseInt64 字符串转int64
func ParseInt64(str string) int64 {
	value, _ := strconv.ParseInt(str, 10, 64)
	return value
}

// ParseUint64 字符串转int64
func ParseUint64(str string) uint64 {
	value, _ := strconv.ParseUint(str, 10, 64)
	return value
}

// Md5Str 计算md5，返回字符串
func Md5Str(str string) string {
	h := md5.New()
	buf := []byte(str)
	count, err := h.Write(buf)
	if err != nil || count != len(buf) {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Md5Byte 计算md5，返回字节
func Md5Byte(str string) []byte {
	h := md5.New()
	buf := []byte(str)
	count, err := h.Write(buf)
	if err != nil || count != len(buf) {
		return nil
	}
	return h.Sum(nil)
}

// Decimal 保留几位小数
func Decimal(value float64, num int) float64 {
	format := "%." + strconv.Itoa(num) + "f"
	value, _ = strconv.ParseFloat(fmt.Sprintf(format, value), 64)
	return value
}

// GetRandomString 生成随机字符串
func GetRandomString(l int) string {
	str := "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

// GetFileNameWithoutSuffix 获取不带后缀文件名
func GetFileNameWithoutSuffix(fullFilename string) string {
	filenameWithSuffix := path.Base(fullFilename)                      //获取文件名带后缀
	fileSuffix := path.Ext(filenameWithSuffix)                         //获取文件后缀
	filenameOnly := strings.TrimSuffix(filenameWithSuffix, fileSuffix) //获取文件名
	return filenameOnly
}

// ToString interface转string
func ToString(value interface{}) string {
	str := ""
	switch v := value.(type) {
	case string:
		str = v
	case int:
		str = strconv.Itoa(v)
	case uint:
		str = strconv.FormatUint(uint64(v), 10)
	case uint8:
		str = strconv.FormatUint(uint64(v), 10)
	case int64:
		str = strconv.FormatInt(v, 10)
	case uint64:
		str = strconv.FormatUint(v, 10)
	case float64:
		str = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		if v {
			str = "true"
		} else {
			str = "false"
		}
	case runtime.Error:
		str = v.Error()
	case error:
		str = v.Error()
	}
	return str
}

// HTTPParam HTTP请求参数
type HTTPParam struct {
	Method   string                 // http请求方法，POST/GET
	URL      string                 // 请求URL
	Data     string                 // 请求数据
	Headers  map[string]interface{} // header
	Cookies  map[string]interface{} // cookie
	UseShort bool                   //使用短链接
	Timeout  uint64                 // 超时设置
}

func createTransport(params *HTTPParam) *http.Transport {
	t := defaultTransport
	if params.UseShort {
		t = &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: true,
		}
	}
	return t
}

// HTTP 发送http请求
func HTTP(params *HTTPParam) ([]byte, int, error) {
	t := createTransport(params)

	var resBody []byte

	body := strings.NewReader(params.Data)
	client := &http.Client{Transport: t, Timeout: time.Second * time.Duration(params.Timeout)}

	params.Method = strings.ToUpper(params.Method)
	if params.Method == "" || params.Method == "GET" {
		params.Method = "GET"
		if !strings.Contains(params.URL, "?") {
			params.URL += "?" + params.Data
		} else {
			params.URL += "&" + params.Data
		}
	}

	request, err := http.NewRequest(params.Method, params.URL, body)
	if err != nil {
		return resBody, 0, err
	}

	// 设置header
	if params.Headers != nil {
		for key, value := range params.Headers {
			strValue := ToString(value)
			request.Header.Set(key, strValue)
		}
	}

	// 设置cookie
	if params.Cookies != nil {
		for key, value := range params.Cookies {
			request.AddCookie(&http.Cookie{Name: key, Value: ToString(value), HttpOnly: true})
		}
	}

	response, err := client.Do(request)
	if err != nil {
		return resBody, 0, err
	}
	defer response.Body.Close()

	resBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return resBody, 0, err
	}
	return resBody, response.StatusCode, nil
}

// GetHTTP 发送http Get请求
func GetHTTP(sURL string) ([]byte, int, error) {
	params := HTTPParam{
		Method: "GET",
		URL:    sURL,
		Data:   "",
	}
	return HTTP(&params)
}

// GeneratePigeonSig 生成pigeon签名
func GeneratePigeonSig(secret string) (string, string) {
	currTime := strconv.FormatInt(time.Now().Unix(), 10)
	h := md5.New()
	value := secret + currTime
	n, err := io.WriteString(h, value)
	if err != nil || n != len(value) {
		return "", ""
	}

	sig := fmt.Sprintf("%x", h.Sum(nil))
	return currTime, sig
}

// CheckLogin 检查和户是否登录
func CheckLogin(nid, token string) (bool, error) {
	clientID := config.Get("client_login", "id").String("")
	secret := config.Get("client_login", "secret").String("")
	currTime, sig := GeneratePigeonSig(secret)
	sURL := config.Get("client_login", "verify_url").String("")

	params := url.Values{}
	params.Set("clientid", clientID)
	params.Set("sig", sig)
	params.Set("timestamp", currTime)
	params.Set("nid", nid)
	params.Set("token", token)
	sURL += "?" + params.Encode()
	body, status, err := GetHTTP(sURL)

	if err != nil {
		return false, err
	}

	if status != 200 {
		return false, fmt.Errorf("Http返回状态码是[%d]。", status)
	}

	result := struct {
		Ret    int64 `json:"ret"`
		Time   int64 `json:"time"`
		PlatID int64 `json:"platid"`
	}{}

	err = json.Unmarshal([]byte(body), &result)
	if err != nil {
		return false, err
	}

	if result.Ret != 0 {
		return false, nil
	}

	return true, nil
}

// CheckFlowAccessLimitLocal 流程访问频率控制
func CheckFlowAccessLimitLocal(nid, actID, flowID string, second, limit int64) bool {
	if second == 0 || limit == 0 {
		return true
	}

	prefix := "ACCESS_" + actID + "_" + flowID + "_" + nid + "_"
	currTime := time.Now().Unix()
	var i int64
	var total int64
	for i = 0; i < second; i++ {
		key := prefix + strconv.FormatInt(currTime-i, 10)
		if count, ok := localCache.Get(key); ok {
			total += count.(int64)
		}
	}
	if total >= limit {
		return false
	}

	redisKey := prefix + strconv.FormatInt(currTime, 10)
	if _, ok := localCache.Get(redisKey); ok {
		_, _ = localCache.IncrementInt64(redisKey, 1)
	} else {
		localCache.Set(redisKey, int64(1), time.Second*time.Duration(second))
	}
	return true
}

// GetIntranetIP 获取当前机器内网IP
func GetIntranetIP() string {
	addrs, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, address := range addrs {
		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// GetEnv 获取当前运行环境
func GetEnv() {
	// 如果编译时没有指定运行环境，则看运行是是否有加“--env=”参数
	env := ""
	flag.StringVar(&env, "env", "", "Running environment.")
	flag.Parse()

	if env == "" && len(os.Args) > 1 {
		env = os.Args[1]
	}

	// 环境配置文件中读到当前运行环境
	envFile := CurrRunPath + "/config/env"
	if _, err := os.Stat(envFile); err == nil {
		buf, err := ioutil.ReadFile(envFile)
		if err == nil {
			env = strings.TrimSpace(string(buf))
		}
	}

	if env != "" {
		Env = env
	}
}

// LoadConfig 载配置文件
func LoadConfig() {
	// 加载通用配置文件
	filepath := CurrRunPath + "/config/"
	configFile := filepath + "config.toml"
	err := config.LoadFile(configFile)
	if err != nil {
		logrus.Fatal(configFile, err)
	}

	configFile = filepath + "db.toml"
	err = config.LoadFile(configFile)
	if err != nil {
		logrus.Fatal(configFile, err)
	}

	configFile = filepath + "message.toml"
	err = config.LoadFile(configFile)
	if err != nil {
		logrus.Fatal(configFile, err)
	}

	// 加载通用配置文件
	envFile := filepath + Env + ".toml"
	_, err = os.Stat(envFile)
	if err == nil {
		err = config.LoadFile(envFile)
		if err != nil {
			logrus.Fatal("读取当前运行环境配置文件" + envFile + "失败。")
		}
	} else {
		logrus.Info("读取当前运行环境配置文件" + envFile + "不存在。")
	}
}

// ShowInfo 显示程序信息
func ShowInfo() {
	fmt.Println("=======================================================================")
	fmt.Println("     Service   : " + config.Get("common", "app_desc").String(""))
	fmt.Println("     Version   : " + Version)
	fmt.Println("     Env       : " + Env)
	fmt.Println("     Commit    : " + Commit)
	fmt.Println("     BuildTime : " + BuildTime)
	fmt.Println("     BuildUser : " + BuildUser)
	fmt.Println("     GoVersion : " + GoVersion)
	fmt.Println("     Address   : " + config.Get("common", "address").String(""))
	fmt.Println("     PProf     : " + config.Get("pprof", "server").String(""))
	fmt.Println("     Metrics   : " + config.Get("monitor", "server").String(""))
	fmt.Println("=======================================================================")
}
