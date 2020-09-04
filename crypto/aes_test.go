package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAes(t *testing.T) {
	key := "1234567890123456"

	value := "hello world."
	buf, err := AesEncrypt([]byte(value), []byte(key))
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	buf, err = AesDecrypt(buf, []byte(key))
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	assert.Equal(t, value, string(buf))
}
