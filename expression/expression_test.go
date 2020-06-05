package expression

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStrlen(t *testing.T) {
	str := "changcai"
	l, err := Strlen(str)
	assert.Nil(t, err)

	count := int(l.(float64))
	assert.Equal(t, count, len(str))

	str = "ch2an2g2ca4iyt4uh1df6gae4rtg8dd0swe34"
	l, err = Strlen(str)
	assert.Nil(t, err)

	count = int(l.(float64))
	assert.Equal(t, len(str), count)
}

func TestIntval(t *testing.T) {
	str := "123"
	tmp, err := Intval(str)
	assert.Nil(t, err)

	val := tmp.(float64)
	assert.Equal(t, float64(123), val)

	str = "1aa"
	tmp, err = Intval(str)
	assert.Nil(t, err)
	assert.Equal(t, float64(0), tmp)
}

func TestDate(t *testing.T) {
	date1 := Date()
	date2 := time.Now().Format("2006-01-02")

	assert.Equal(t, date1, date2)
}

func TestTime(t *testing.T) {
	s := Time()
	assert.NotEmpty(t, s)
}

func TestEval(t *testing.T) {
	str := "12 > 4 && (54 - 5) / 6 < 7"
	tmp, err := Eval(str)
	assert.Nil(t, err)

	result, ok := tmp.(bool)
	assert.True(t, ok)
	assert.Equal(t, (12 > 4 && (54-5)/6 < 7), result)

	str = "765 / 5"
	tmp, err = Eval(str)
	assert.Nil(t, err)

	resultB, ok := tmp.(float64)
	assert.True(t, ok)
	assert.Equal(t, float64(765/5), resultB)

	str = `strlen("changcai")`
	tmp, err = Eval(str)
	assert.Nil(t, err)

	resultC, ok := tmp.(float64)
	assert.True(t, ok)
	assert.Equal(t, len("changcai"), int(resultC))

	str = `intval("345")`
	tmp, err = Eval(str)
	assert.Nil(t, err)

	resultD, ok := tmp.(float64)
	assert.True(t, ok)
	assert.Equal(t, 345, int(resultD))

	str = `"12" fd> || 4 && (54 - 5) / 6 < 7`
	_, err = Eval(str)
	assert.NotNil(t, err)
}
