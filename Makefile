default: dev

PACKAGE="github.com/yuanzhangcai/chaos/common"

USER=`whoami`
GIT_TAG=`git describe --tags`
GIT_COMMIT=`git rev-parse HEAD`
BUILD_TIME=`date '+%F %T'`
VERSION=`git rev-list --tags --max-count=1 | xargs git describe --tags`
GO_VERSION=`go version`
LDFLAGS="-X ${PACKAGE}.Commit=${GIT_COMMIT} -X '${PACKAGE}.BuildTime=${BUILD_TIME}' -X ${PACKAGE}.BuildUser=${USER} -X ${PACKAGE}.Version=${VERSION} -X '${PACKAGE}.GoVersion=${GO_VERSION}' "

define prepare
	#go-bindata -o ./bindata/bindata.go -pkg bindata -fs  html/*
endef

# 执行该命令会给go.mod添加gowatch、go-bindata相关的依赖包，建议到工作目录外手动执行go get
install:
	go get github.com/silenceper/gowatch
	#go get -u github.com/go-bindata/go-bindata/...

# 热编译启动程序 默认指定运行环境为本地运行环境，配置详见gowatch.yml
watch:
	$(call prepare)
	gowatch

# 编译本地开发运行环境程序
dev:
	$(call prepare)
	go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=dev" -race

# 编译测试运行环境程序
test:
	$(call prepare)
	go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=test" -race

# 编译预发布环境程序
pre:
	$(call prepare)
	go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=pre" -race

# 编译正式运行环境程序
prod:
	$(call prepare)
	go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=prod"
	# CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=prod" -o chaos_macOS
	# CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=prod" -o chaos_linux
	# CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=prod" -o chaos_windows.exe

# sonar代码扫描
sonar:
	$(call prepare)
	ulimit -n 24000
	go vet -n ./... 2> ./vet.tmp
	golangci-lint run ./... --out-format=checkstyle > golangci-lint.tmp || true
	go test -race -cover -v  ./... -json -coverprofile=covprofile > test.tmp
	sonar-scanner \
	-Dsonar.host.url=http://127.0.0.1:9000 \
	-Dsonar.sources=. \
	-Dsonar.tests=. \
	-Dsonar.exclusions="**/*_test.go" \
	-Dsonar.projectKey=chaos \
	-Dsonar.login=344ad0c611674bcbbf571f17bf5271f4c678e4aa \
	-Dsonar.go.tests.reportPaths=test.tmp \
	-Dsonar.go.coverage.reportPaths=covprofile \
	-Dsonar.go.govet.reportPaths=vet.tmp \
	-Dsonar.go.golangci-lint.reportPaths=golangci-lint.tmp \
	-Dsonar.test.inclusions="**/*_test.go" \
    -Dsonar.test.exclusions="**/vendor/**" | grep -v "WARN:"
	rm -rf *.tmp
	rm -rf .scannerwork
	rm -rf covprofile


# 编译并生成镜像文件
image:
	$(call prepare)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags ${LDFLAGS}" -X ${PACKAGE}.Env=prod"
	docker build  -t chaos .

.PHONY: install watch dev test pre prod image