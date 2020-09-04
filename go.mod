module github.com/yuanzhangcai/chaos

go 1.15

replace github.com/coreos/go-systemd => github.com/coreos/go-systemd/v22 v22.1.0

require (
	github.com/Knetic/govaluate v3.0.0+incompatible
	github.com/gin-gonic/gin v1.6.3
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/jinzhu/gorm v1.9.16
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/prometheus/client_golang v1.7.1
	github.com/robfig/cron v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.1
	github.com/yuanzhangcai/config v0.0.0-20200806074344-66e1e22e6731
	github.com/yuanzhangcai/srsd v0.0.0-20200819035745-0388399ef1ba
)
