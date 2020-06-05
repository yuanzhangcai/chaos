package tools

import (
	"testing"
	"time"

	"github.com/micro/go-micro/v2/client/selector"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	"github.com/stretchr/testify/assert"
)

var etcdAddress = []string{"47.99.79.44:2379", "47.111.108.59:2379", "47.99.62.229:2379"}

func TestRegister(t *testing.T) {
	opt := &RegisterOptions{}
	r1 := NewServicesRegister(opt)
	assert.NotNil(t, r1)

	r2 := NewServicesRegister(opt)
	assert.NotNil(t, r2)

	assert.NotEqual(t, r1, r2)

	serverName := "test.chaos.zacyuan.com." + time.Now().String()
	opt = &RegisterOptions{
		ServerName:    serverName,
		EtcdAddress:   etcdAddress,
		ServerAddress: "",
		TTL:           2,
		Interval:      1,
	}
	register := NewServicesRegister(opt)
	err := register.Start()
	assert.NotNil(t, err)

	opt.ServerAddress = ":4004"
	register = NewServicesRegister(opt)
	err = register.Start()
	assert.Nil(t, err)

	Selector := selector.NewSelector(
		selector.Registry(
			etcd.NewRegistry(func(op *registry.Options) {
				op.Addrs = etcdAddress
			}),
		),
	)

	_, err = Selector.Select(serverName)
	assert.Nil(t, err)

	_ = register.Stop()
	time.Sleep(3 * time.Second)
	_, err = Selector.Select(serverName)
	assert.NotNil(t, err)

}
