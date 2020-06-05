package tools

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/registry/etcd"
	maddr "github.com/micro/go-micro/v2/util/addr"
)

// RegisterOptions 服注册参数
type RegisterOptions struct {
	ServerName    string
	EtcdAddress   []string
	Version       string
	ServerAddress string
	TTL           time.Duration // 服务失效时间
	Interval      time.Duration // 服务注册间隔时间
}

// ServicesRegister 服务注册组件
type ServicesRegister struct {
	opt  *RegisterOptions  // 服务注册参数
	r    registry.Registry // 服务注册器
	svr  *registry.Service // 服务注册信息
	exit chan int          // 限出服务注册
}

// NewServicesRegister 创建注册器
func NewServicesRegister(opt *RegisterOptions) *ServicesRegister {
	return &ServicesRegister{
		opt:  opt,
		exit: make(chan int),
	}
}

// Start 开启服务注册
func (c *ServicesRegister) Start() error {
	c.r = etcd.NewRegistry(func(op *registry.Options) {
		op.Addrs = c.opt.EtcdAddress
	})

	host, port, err := net.SplitHostPort(c.opt.ServerAddress)
	if err != nil {
		return err
	}

	addr, err := maddr.Extract(host)
	if err != nil {
		return err
	}

	if strings.Count(addr, ":") > 0 {
		addr = "[" + addr + "]"
	}

	if c.opt.Version == "" {
		c.opt.Version = "latest"
	}

	c.svr = &registry.Service{
		Name:    c.opt.ServerName,
		Version: c.opt.Version,
		Nodes: []*registry.Node{{
			Id:       uuid.New().String(),
			Address:  fmt.Sprintf("%s:%s", addr, port),
			Metadata: nil,
		}},
	}

	err = c.r.Register(c.svr, registry.RegisterTTL(c.opt.TTL))
	if err != nil {
		return err
	}

	// 启动定时刷新功能
	go c.run()

	return nil
}

// Stop 停止服务注册
func (c *ServicesRegister) Stop() error {
	close(c.exit) // 关闭定时刷新功能

	if c.svr == nil {
		return nil
	}

	return c.r.Deregister(c.svr)
}

// run 定时注册服务信息
func (c *ServicesRegister) run() {
	if c.opt.Interval <= time.Duration(0) {
		return
	}

	t := time.NewTicker(c.opt.Interval)

	for {
		select {
		case <-t.C:
			_ = c.r.Register(c.svr, registry.RegisterTTL(c.opt.TTL))
		case <-c.exit:
			t.Stop()
			return
		}
	}
}
