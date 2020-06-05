package common

// 日期格式化模版
const (
	Y       string = "2006"
	YM      string = "2006-01"
	YMD     string = "2006-01-02"
	YMD2    string = "20060102"
	YMDH    string = "2006-01-02 15"
	YMDHI   string = "2006-01-02 15:04"
	YMDHI2  string = "200601021504"
	YMDHIS  string = "2006-01-02 15:04:05"
	YMDHIS2 string = "20060102150405"
	HI      string = "15:04"
	HI2     string = "1504"
)

var (
	// CurrRunPath 当前运行程序所在目录
	CurrRunPath string

	// CurrRunFileName 当前运行程序文件名
	CurrRunFileName string
)
