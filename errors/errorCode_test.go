package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New(0, "OK")
	assert.NotNil(t, err)
	assert.Equal(t, int64(0), err.Code)
	assert.Equal(t, "OK", err.Msg)

	err2 := New(-999, "系统错误。")
	assert.NotEqual(t, err, err2)
}

func TestError(t *testing.T) {
	err := New(0, "OK")
	assert.Equal(t, "OK", err.Error())
}
