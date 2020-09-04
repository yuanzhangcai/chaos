package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRsa(t *testing.T) {
	err := GenerateRSAKey(256, "")
	assert.Nil(t, err)

	value := "hello world."
	buf, err := RsaEncryptByFile([]byte(value), "./public.pem")
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	buf, err = RsaDecryptByFile(buf, "./private.pem")
	assert.Nil(t, err)
	assert.NotNil(t, buf)

	assert.Equal(t, value, string(buf))
}
