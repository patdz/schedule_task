package schedule_task

import (
	"errors"
	"github.com/satori/go.uuid"
	"sync"
	"time"
)

const (
	taskOpAdd = iota
	taskOpDel
	taskOpUpdate
)

type task struct {
	cycleNum int
	exec     TaskFunc
	params   []interface{}
}

type taskOpInfo struct {
	opType int
	key    string
	at     time.Time
	exec   TaskFunc
	params []interface{}
}

// TimeWheelTaskPool .
type TimeWheelTaskPool struct {
	sync.WaitGroup
	curIndex  int64
	closed    chan bool
	taskOp    chan *taskOpInfo
	startTime time.Time
	slots     [3600]map[string]*task
	keys      map[string]int
}

// NewTimeWheelTaskPool .
func NewTimeWheelTaskPool() *TimeWheelTaskPool {
	pool := &TimeWheelTaskPool{}
	pool.Start()
	return pool
}

func (tp *TimeWheelTaskPool) Start() {
	tp.Wait()
	tp.curIndex = 0
	tp.closed = make(chan bool)
	tp.taskOp = make(chan *taskOpInfo, 100)
	tp.startTime = time.Now()
	for i := 0; i < 3600; i++ {
		tp.slots[i] = make(map[string]*task)
	}
	tp.keys = make(map[string]int)

	tp.Add(1)
	go tp.timeLoop()
}

func (tp *TimeWheelTaskPool) Stop() {
	close(tp.closed)
	tp.Wait()
}

func (tp *TimeWheelTaskPool) execSlot() {
	tasks := tp.slots[tp.curIndex%3600]
	if len(tasks) <= 0 {
		return
	}
	for k, v := range tasks {
		if v.cycleNum == 0 {
			tp.execTask(v.exec, v.params)
			tp.delTask(k)
		} else {
			v.cycleNum--
		}
	}
}

func (tp *TimeWheelTaskPool) execTask(exec TaskFunc, params []interface{}) {
	go exec(params...)
}

func (tp *TimeWheelTaskPool) addOrUpdateTask(t time.Time, key string, exec TaskFunc, params []interface{}) {
	if tp.startTime.Unix()+tp.curIndex >= t.Unix() {
		tp.execTask(exec, params)
		return
	}
	if key == "" {
		key = uuid.NewV4().String()
	} else {
		tp.delTask(key)
	}
	subSecond := t.Unix() - tp.startTime.Unix()
	cycleNum := int(subSecond / 3600)
	idx := subSecond % 3600
	tp.slots[idx][key] = &task{
		cycleNum: cycleNum,
		exec:     exec,
		params:   params,
	}
	tp.keys[key] = int(idx)
}

func (tp *TimeWheelTaskPool) delTask(key string) {
	if idx, ok := tp.keys[key]; ok {
		delete(tp.slots[idx], key)
		delete(tp.keys, key)
	}
}

func (tp *TimeWheelTaskPool) timeLoop() {
	defer tp.Done()

	tick := time.NewTicker(time.Second)
	for {
		select {
		case <-tp.closed:
			return
		case taskI := <-tp.taskOp:
			if taskI.opType == taskOpAdd || taskI.opType == taskOpUpdate {
				tp.addOrUpdateTask(taskI.at, taskI.key, taskI.exec, taskI.params)
			} else if taskI.opType == taskOpDel {
				tp.delTask(taskI.key)
			}
		case <-tick.C:
			tp.curIndex++
			tp.execSlot()
		}
	}
}

func (tp *TimeWheelTaskPool) AddTask(key string, t time.Time, exec TaskFunc, params ...interface{}) error {
	if exec == nil {
		return errors.New("exec is nil")
	}
	tp.taskOp <- &taskOpInfo{
		opType: taskOpAdd,
		key:    key,
		at:     t,
		exec:   exec,
		params: params,
	}

	return nil
}

func (tp *TimeWheelTaskPool) RemoveTask(key string) error {
	if key != "" {
		tp.taskOp <- &taskOpInfo{
			opType: taskOpDel,
			key:    key,
		}
	}
	return nil
}
