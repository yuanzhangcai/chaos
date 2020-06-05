package tools

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	server   = "10.10.40.49:6379"
	password = "12345678"
	prefix   = ""
)

func TestInitRedis(t *testing.T) {
	err := InitRedis("", password, prefix)
	assert.NotNil(t, err)

	err = InitRedis(server+"1", password, prefix)
	assert.NotNil(t, err)

	// 初次初始化
	err = InitRedis(server, password, prefix)
	assert.Nil(t, err)

	// 重复初始化
	err = InitRedis(server, password, prefix)
	assert.Nil(t, err)
}

func TestGetRedis(t *testing.T) {
	assert.Equal(t, client, GetRedis())
}

func TestCommand(t *testing.T) {
	_ = InitRedis(server, password, prefix)

	valStr := "SetObject"
	err := client.SetObject("test_set_string", valStr, 300*time.Second)
	assert.Nil(t, err)
	newValStr := ""
	err = client.GetObject("test_set_string", &newValStr)
	assert.Nil(t, err)
	assert.Equal(t, valStr, newValStr)

	valInt := 453
	err = client.SetObject("test_set_int", valInt, 300*time.Second)
	assert.Nil(t, err)
	newValInt := 0
	err = client.GetObject("test_set_int", &newValInt)
	assert.Nil(t, err)
	assert.Equal(t, valInt, newValInt)

	valMap := map[string]interface{}{"name": "zacyuan"}
	err = client.SetObject("test_set_map", valMap, 300*time.Second)
	assert.Nil(t, err)
	newValMap := make(map[string]interface{})
	err = client.GetObject("test_set_map", &newValMap)
	assert.Nil(t, err)
	assert.Equal(t, valMap["name"], newValMap["name"])

	type testStruct struct {
		Name string
		Age  int
	}
	valStc := testStruct{Name: "zacyuan", Age: 18}
	err = client.SetObject("test_set_struct", valStc, 300*time.Second)
	assert.Nil(t, err)
	newValStc := testStruct{}
	err = client.GetObject("test_set_struct", &newValStc)
	assert.Nil(t, err)
	assert.Equal(t, valStc.Name, newValStc.Name)
	assert.Equal(t, valStc.Age, newValStc.Age)

	newValStc2 := testStruct{}
	err = client.GetObject("test_set_struct2", &newValStc2)
	assert.NotNil(t, err)

	key := client.GenerateScoreKey(10, "123")
	assert.Equal(t, client.prefix+"10_123", key)

	score := map[string]int{"score": 111}

	err = client.ScoreSet(10, "123", score, 300*time.Second)
	assert.Nil(t, err)

	scoreNew := make(map[string]int)
	err = client.ScoreGet(10, "123", &scoreNew)
	assert.Nil(t, err)
	assert.Equal(t, score["score"], scoreNew["score"])

	ret := client.ScoreDel(10, "123")
	assert.Nil(t, ret.Err())
	assert.Equal(t, int64(1), ret.Val())

	scoreNew2 := make(map[string]int)
	err = client.ScoreGet(10, "123", &scoreNew2)
	assert.NotNil(t, err)

	ret = client.ScoreDel(10, "123")
	assert.Nil(t, ret.Err())
	assert.Equal(t, int64(0), ret.Val())
}
