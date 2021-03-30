package tools

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
)

// Redis redis组件
type Redis struct {
	*redis.Client
	prefix string
}

var client *Redis

// NewRedis 创建redis对象
func NewRedis(server, password, prefix string) (*Redis, error) {
	if server == "" {
		return nil, errors.New("redis服务器地址为空。")
	}

	cli := &Redis{
		Client: redis.NewClient(&redis.Options{
			Addr:     server,
			Password: password,
		}),
		prefix: prefix,
	}

	pong, err := cli.Ping().Result()
	if err != nil {
		logrus.Error(pong)
		logrus.Error(err)
		return nil, err
	}
	return cli, nil
}

// InitRedis 初始化redis
func InitRedis(server, password, prefix string) error {
	if client != nil {
		return nil
	}

	cli, err := NewRedis(server, password, prefix)
	if err != nil {
		return err
	}

	client = cli
	return nil
}

// GetRedis 获取redis实例
func GetRedis() *Redis {
	return client
}

// SetObject 设置redis对象
func (c *Redis) SetObject(key string, value interface{}, expire time.Duration) error {
	key = c.prefix + key
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Set(key, string(buf), expire).Err()
}

// GetObject 获取redis对象
func (c *Redis) GetObject(key string, value interface{}) error {
	key = c.prefix + key
	ret := c.Get(key)
	if ret.Err() != nil {
		return ret.Err()
	}

	err := json.Unmarshal([]byte(ret.Val()), value)
	if err == nil {
		return err
	}

	return nil
}

// GenerateScoreKey 生成积分redis的key
func (c *Redis) GenerateScoreKey(scoreID uint64, nid string) string {
	return c.prefix + strconv.FormatUint(scoreID, 10) + "_" + nid
}

// ScoreSet 写入积分redis
func (c *Redis) ScoreSet(scoreID uint64, nid string, value interface{}, expire time.Duration) error {
	key := c.GenerateScoreKey(scoreID, nid)
	buf, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.Set(key, string(buf), expire).Err()
}

// ScoreGet 获取积分redis
func (c *Redis) ScoreGet(scoreID uint64, nid string, value interface{}) error {
	key := c.GenerateScoreKey(scoreID, nid)
	ret := c.Get(key)
	if ret.Err() != nil {
		return ret.Err()
	}

	err := json.Unmarshal([]byte(ret.Val()), value)
	if err != nil {
		return err
	}

	return nil
}

// ScoreDel 删除积分redis
func (c *Redis) ScoreDel(scoreID uint64, nid string) *redis.IntCmd {
	key := c.GenerateScoreKey(scoreID, nid)
	return c.Del(key)
}
