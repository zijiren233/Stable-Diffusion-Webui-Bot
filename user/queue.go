package user

import (
	"errors"
	"fmt"
	"sync"

	parseflag "github.com/zijiren233/stable-diffusion-webui-bot/flag"
)

var queue = &sync.Map{}

type TaskType uint

const (
	T_Draw TaskType = iota
	T_GuessTag
	T_CtrlPhoto
	T_SuperResolution
)

type Task struct {
	id       int64
	taskType TaskType
	values   map[string]any
	// cancel   context.CancelFunc
	// result   <-chan *ai.Resoult
	// status   func() ai.Status
}

func (t *Task) Set(key string, value any) {
	t.values[key] = value
}

func (t *Task) Value(key string) any {
	return t.values[key]
}

func GetTask(userID int64, types TaskType) (*Task, error) {
	value, ok := queue.Load(fmt.Sprintf("%d:%d", userID, types))
	if ok {
		return value.(*Task), nil
	} else {
		return nil, errors.New("task not fount")
	}
}

func (u *UserInfo) AddTask(types TaskType) (*Task, error) {
	if parseflag.Dev && u.UserInfo.UserID != parseflag.MyID {
		return nil, errors.New("server under maintenance, please try again later")
	}
	task := &Task{id: u.UserInfo.UserID, taskType: types}
	if actual, ok := queue.LoadOrStore(fmt.Sprintf("%d:%d", u.UserInfo.UserID, types), task); ok {
		return actual.(*Task), errors.New("task already exists")
	} else {
		task.values = make(map[string]any)
		return task, nil
	}
}

func (task *Task) Down() {
	queue.Delete(fmt.Sprintf("%d:%d", task.id, task.taskType))
}

func (task *Task) ID() int64 {
	return task.id
}

func (task *Task) Type() TaskType {
	return task.taskType
}
