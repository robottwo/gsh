package filesystem

import "os"

type FileSystem interface {
	Open(name string) (*os.File, error)
	Create(name string) (*os.File, error)
}

type DefaultFileSystem struct{}

func (fs DefaultFileSystem) Open(name string) (*os.File, error) {
	return os.Open(name)
}

func (fs DefaultFileSystem) Create(name string) (*os.File, error) {
	return os.Create(name)
}
