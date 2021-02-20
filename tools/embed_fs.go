package tools

import (
	"embed"
	"io/fs"
	"path"
)

// EmbedFS 静态资源
type EmbedFS struct {
	// embed静态资源
	fs embed.FS
	// 设置embed文件到静态资源的相对路径，也就是embed注释里的路径
	path string
}

// NewEmbedFS 创建EmbedFS对象
func NewEmbedFS(fs embed.FS, path string) *EmbedFS {
	return &EmbedFS{
		fs:   fs,
		path: path,
	}
}

// Open 打开文件
func (c *EmbedFS) Open(name string) (fs.File, error) {
	var fullName string
	fullName = path.Join(c.path, name)
	return c.fs.Open(fullName)
}

// ReadDir 读取目录
func (c *EmbedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var fullName string
	fullName = path.Join(c.path, name)
	return c.fs.ReadDir(fullName)
}

// ReadFile 读取文件
func (c *EmbedFS) ReadFile(name string) ([]byte, error) {
	var fullName string
	fullName = path.Join(c.path, name)
	return c.fs.ReadFile(fullName)
}
