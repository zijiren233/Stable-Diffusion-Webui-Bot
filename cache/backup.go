package cache

import (
	"io/fs"
)

type FileInfo struct {
	FileID   string
	FilePath string
	Info     fs.FileInfo
}

type Cache interface {
	Put(data []byte) (info FileInfo, err error)
	Get(md5 string) (data []byte, err error)
	GetStat(md5 string) (fs.FileInfo, error)
}
