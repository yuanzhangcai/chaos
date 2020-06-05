package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersion(t *testing.T) {
	Version = "V0.0.1"
	Env = "prod"
	Commit = "802ae0c32427ad3cab9faff0cc468bb32e19d236"
	BuildTime = "2020-04-30 04:02:39"
	BuildUser = "jenkins"
	GoVersion = "go version go1.14 linux/amd64"

	info := GetVersion()
	assert.Equal(t, Version, info["version"])
	assert.Equal(t, Env, info["env"])
	assert.Equal(t, Commit, info["commit"])
	assert.Equal(t, BuildTime, info["build_time"])
	assert.Equal(t, BuildUser, info["build_user"])
	assert.Equal(t, GoVersion, info["go_version"])
}
