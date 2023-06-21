package gconfig

import (
	"context"
	"os"
	"sync"

	"github.com/OhYee/rainbow/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/zijiren233/go-colorlog"
)

type watchs struct {
	ctx    context.Context
	cancel context.CancelFunc
	ch     chan struct{}
}

func (w *watchs) Ch() <-chan struct{} {
	return w.ch
}

func (w *watchs) Close() {
	w.cancel()
}

var (
	watchMap = &sync.Map{}
)

func NewWatchConfig(ctx context.Context) *watchs {
	ch := make(chan struct{}, 1)
	ctx, cancel := context.WithCancel(ctx)
	w := &watchs{ctx: ctx, cancel: cancel, ch: ch}
	watchMap.Store(w, w)
	return w
}

func watch(file string) {
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		colorlog.Errorf(errors.ShowStack(err))
		return
	}
	defer watch.Close()
	f, err := os.OpenFile(file, os.O_RDONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		colorlog.Errorf(errors.ShowStack(err))
		return
	}
	defer f.Close()
	err = watch.Add(file)
	if err != nil {
		colorlog.Errorf(errors.ShowStack(err))
		return
	}
	for {
		select {
		case ev, ok := <-watch.Events:
			if !ok {
				return
			}
			if ev.Op.Has(fsnotify.Write) {
				if err := Load(ConfigPath); err != nil {
					colorlog.Errorf("Load config err: %v", err)
				} else {
					colorlog.Debugf("Load config success: %v", config)
					watchMap.Range(func(key, value any) bool {
						v := value.(*watchs)
						select {
						case <-v.ctx.Done():
							close(v.ch)
							watchMap.Delete(key)
						default:
							v.ch <- struct{}{}
						}
						return true
					})
				}
			}
			// if ev.Op&fsnotify.Remove == fsnotify.Remove {
			// 	log.Debug.Println("删除文件 : ", ev.Name)
			// }
			// if ev.Op&fsnotify.Rename == fsnotify.Rename {
			// 	log.Debug.Println("重命名文件 : ", ev.Name)
			// }
			// if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
			// 	log.Debug.Println("修改权限 : ", ev.Name)
			// }
		case err := <-watch.Errors:
			colorlog.Errorf(errors.ShowStack(err))
		}
	}
}
