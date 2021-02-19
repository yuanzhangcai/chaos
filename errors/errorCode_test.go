package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	DBErr    = New(-800, "数据库操作失败")
	DBQual   = New(-700, "资格操作失败")
	RedisErr = New(-600, "Redis error")
)

func TestNew(t *testing.T) {
	err := New(0, "OK")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), err.Code())
	assert.Equal(t, "OK", err.Msg())

	err2 := New(-999, "系统错误。")
	assert.NotEqual(t, err, err2)
}

func TestError(t *testing.T) {
	err := New(0, "OK")
	assert.Equal(t, "OK", err.Error())
}

func TestAs(t *testing.T) {
	e1 := fmt.Errorf("insert error")
	e2 := Wrap(DBErr, e1)
	e3 := Wrap(DBQual, e2)

	assert.True(t, e3.As(e1))
	assert.True(t, e3.As(e2))
	assert.True(t, e3.As(e3))
	assert.False(t, e3.As(RedisErr))
	assert.False(t, e3.As(nil))

	var e *Error
	assert.True(t, e.As(nil))

}

func TestWrap(t *testing.T) {
	e1 := fmt.Errorf("db error")
	e2 := Wrap(DBErr, e1)
	e3 := Wrap(DBQual, e2)

	assert.Equal(t, e2.Code(), DBErr.Code())
	assert.Equal(t, "数据库操作失败 -> db error", e2.Error())

	assert.Equal(t, e3.Code(), DBQual.Code())
	assert.Equal(t, "资格操作失败 -> 数据库操作失败 -> db error", e3.Error())
}

func TestWrapStr(t *testing.T) {
	e1 := WrapStr(nil, "db error")
	assert.Equal(t, int64(0), e1.Code())
	assert.Equal(t, "db error", e1.Error())

	e2 := WrapStr(e1, "time out")
	assert.Equal(t, int64(0), e2.Code())
	assert.Equal(t, "db error -> time out", e2.Error())
}

func TestCause(t *testing.T) {
	e1 := fmt.Errorf("db error")
	e2 := Wrap(DBErr, e1)
	e3 := Wrap(DBQual, e2)

	assert.Equal(t, e1, Cause(e3))
}
