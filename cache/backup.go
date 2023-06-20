package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/zijiren233/stable-diffusion-webui-bot/utils"

	"github.com/bluele/gcache"
	"github.com/im7mortal/kmutex"
	"golang.org/x/time/rate"
)

var (
	Path      = `/mnt/stable-diffusion-webui-bot`
	fileLock  = &sync.Mutex{}
	fileLimit = rate.NewLimiter(9, 1)
	kFileLock = kmutex.New()
)

var (
	errDataIsNil = errors.New("data is nil")
)

// userid -> langtype
var fileID2Byte = gcache.New(3000).LRU().Build()

type FileInfo struct {
	Md5      string
	FilePath string
	Info     fs.FileInfo
}

func Put(data []byte) (info FileInfo, err error) {
	if data == nil {
		err = errDataIsNil
		return
	}
	info.Md5 = utils.Md5(data)
	info.FilePath = md52Path(info.Md5)
	fileLimit.Wait(context.Background())
	fileLock.Lock()
	info.Info, err = os.Stat(info.FilePath)
	if err == nil && info.Info.Size() == int64(len(data)) {
		fileLock.Unlock()
		return
	}
	err = os.MkdirAll(path.Dir(info.FilePath), os.ModePerm)
	fileLock.Unlock()
	if err != nil {
		return
	}
	kFileLock.Lock(info.Md5)
	defer kFileLock.Locker(info.Md5)
	err = os.WriteFile(info.FilePath, data, os.ModePerm)
	if err != nil {
		return
	}
	SetFile(info.Md5, data)
	return
}

var errMd5Len = errors.New("md5 len err")

func GetFile(md5 string) (data []byte, err error) {
	l := len(md5)
	if l == 0 {
		return
	} else if l != 32 {
		return nil, errMd5Len
	}
	i, err := fileID2Byte.Get(md5)
	if err != nil {
		fileLimit.Wait(context.Background())
		data, err = os.ReadFile(md52Path(md5))
		if err != nil {
			return
		}
		SetFile(md5, data)
	} else {
		data = i.([]byte)
	}
	return
}

func GetFileStat(md5 string) (fs.FileInfo, error) {
	l := len(md5)
	if l == 0 {
		return nil, nil
	} else if l != 32 {
		return nil, errMd5Len
	}
	return os.Stat(md52Path(md5))
}

func md52Path(md5 string) string {
	builder := strings.Builder{}
	builder.WriteString(Path)
	for k, v := range md5 {
		if k == 4 {
			break
		}
		builder.WriteRune('/')
		builder.WriteRune(v)
	}
	return fmt.Sprintf("%s/%s.png", builder.String(), md5)
}

func SetFile(md5 string, data []byte) {
	fileID2Byte.SetWithExpire(md5, data, time.Minute*30)
}
