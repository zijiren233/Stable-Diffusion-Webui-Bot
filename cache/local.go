package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"time"

	"github.com/bluele/gcache"
	"github.com/zijiren233/ksync"
	"github.com/zijiren233/stable-diffusion-webui-bot/utils"
	"golang.org/x/time/rate"
)

type LocalCache struct {
	fileLimit *rate.Limiter
	kFileLock *ksync.Kmutex
	savePath  string
	cacheNum  int
	readLimit int
	cache     gcache.Cache
}

type LocalCacheFunc func(l *LocalCache)

func WithSavePath(savePath string) LocalCacheFunc {
	return func(l *LocalCache) { l.savePath = savePath }
}

func WithCacheNum(cacheNum int) LocalCacheFunc {
	return func(l *LocalCache) { l.cacheNum = cacheNum }
}

func WithReatLimit(readLimit int) LocalCacheFunc {
	return func(l *LocalCache) { l.readLimit = readLimit }
}

func NewCache(configs ...LocalCacheFunc) (Cache, error) {
	lc := new(LocalCache)
	for _, c := range configs {
		c(lc)
	}
	if lc.savePath == "" {
		lc.savePath = path.Join(os.TempDir(), "local-cache")
	}
	if err := lc.init(); err != nil {
		return nil, err
	}
	if lc.readLimit <= 0 {
		lc.readLimit = 9
	}
	if lc.cacheNum <= 0 {
		lc.cacheNum = 1
	}
	lc.fileLimit = rate.NewLimiter(rate.Limit(lc.readLimit), 1)
	lc.kFileLock = ksync.NewKmutex()
	lc.cache = gcache.New(lc.cacheNum).LRU().Build()
	return lc, nil
}

var (
	errDataIsNil = errors.New("data is nil")
)

func (l *LocalCache) init() error {
	return os.MkdirAll(l.savePath, os.ModePerm)
}

// userid -> langtype
func (l *LocalCache) Put(data []byte) (info FileInfo, err error) {
	if data == nil {
		err = errDataIsNil
		return
	}
	id := utils.Md5(data)
	info.FileID = id
	info.FilePath = l.id2Path(id)
	l.fileLimit.Wait(context.Background())
	l.kFileLock.Lock(id)
	defer l.kFileLock.Unlock(id)
	info.Info, err = os.Stat(info.FilePath)
	if err == nil && info.Info.Size() == int64(len(data)) {
		return
	}
	err = os.MkdirAll(path.Dir(info.FilePath), os.ModePerm)
	if err != nil {
		return
	}
	err = os.WriteFile(info.FilePath, data, os.ModePerm)
	if err != nil {
		return
	}
	l.SetFile(id, data)
	return
}

var errIDLen = errors.New("id len err")

func (l *LocalCache) Get(id string) (data []byte, err error) {
	if len(id) != 32 {
		return nil, errIDLen
	}
	i, err := l.cache.Get(id)
	if err != nil {
		l.fileLimit.Wait(context.Background())
		data, err = os.ReadFile(l.id2Path(id))
		if err != nil {
			return
		}
		l.SetFile(id, data)
	} else {
		data = i.([]byte)
	}
	return
}

func (l *LocalCache) GetStat(id string) (fs.FileInfo, error) {
	if len(id) != 32 {
		return nil, errIDLen
	}
	return os.Stat(l.id2Path(id))
}

func (l *LocalCache) id2Path(id string) string {
	builder := strings.Builder{}
	builder.WriteString(l.savePath)
	for k, v := range id {
		if k == 4 {
			break
		}
		builder.WriteRune('/')
		builder.WriteRune(v)
	}
	return fmt.Sprintf("%s/%s.png", builder.String(), id)
}

func (l *LocalCache) SetFile(md5 string, data []byte) {
	l.cache.SetWithExpire(md5, data, time.Minute*30)
}
