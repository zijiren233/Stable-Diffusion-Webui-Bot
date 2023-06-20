package gconfig

import (
	"os"

	"github.com/OhYee/rainbow/errors"
	"github.com/fsnotify/fsnotify"
	"github.com/zijiren233/go-colorlog"
)

var watchList = []chan struct{}{}

func NewWatchConfig() <-chan struct{} {
	ch := make(chan struct{}, 1)
	watchList = append(watchList, ch)
	return ch
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
				err := Load(ConfigPath)
				if err != nil {
					colorlog.Errorf("Load config err: %v", err)
				} else {
					colorlog.Debugf("Load config success: %v", config)
					for _, v := range watchList {
						v <- struct{}{}
					}
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
