package errors

// Error 异常类型
type Error struct {
	code  int64  // 错误码
	msg   string // 错误信息
	cause error  // error
}

func (c *Error) Error() string {
	if c.cause != nil {
		return c.msg + " -> " + c.cause.Error()
	}
	return c.msg
}

// Code 返回错误码
func (c *Error) Code() int64 {
	return c.code
}

// Msg 返回错误信息
func (c *Error) Msg() string {
	return c.msg
}

// As 判断该错误是否是指定错误
func (c *Error) As(err error) bool {
	if err == nil {
		return c == nil
	}

	var e error = c
	for e != nil {
		if tmp, ok := e.(*Error); ok {
			if as, ok := err.(*Error); ok {
				if tmp.Code() == as.Code() {
					return true
				}
			}
			e = tmp.cause
		} else {
			if e == err || e.Error() == err.Error() {
				return true
			}
			return false
		}
	}

	return false
}

// New 创建错误
func New(code int64, msg string) *Error {
	return &Error{
		code: code,
		msg:  msg,
	}
}

// Wrap Wrap
func Wrap(err *Error, cause error) *Error {
	if err == nil {
		return &Error{cause: cause}
	}

	return &Error{
		code:  err.code,
		msg:   err.msg,
		cause: cause,
	}
}

// WrapStr WrapStr
func WrapStr(err *Error, msg string) *Error {
	if err == nil {
		return &Error{msg: msg}
	}
	cause := New(0, msg)
	return Wrap(err, cause)
}

// Cause Cause
func Cause(err *Error) error {
	for err != nil {
		if err.cause == nil {
			break
		}

		if cause, ok := err.cause.(*Error); ok {
			err = cause
		} else {
			return err.cause
		}
	}
	return err
}

var (
	// OK 正常
	OK = New(0, "OK")

	// ErrSystem 系统错误
	ErrSystem = New(-9999, "系统错误")
)
