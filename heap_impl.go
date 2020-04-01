package schedule_task

import (
	"container/heap"
	"errors"
	uuid "github.com/satori/go.uuid"
	"sync"
	"time"
)

type taskItem struct {
	at        time.Time
	exec      TaskFunc
	params    []interface{}
	cancelled bool
	key       string
}

type taskHeap []*taskItem

func (th taskHeap) Len() int {
	return len(th)
}

func (th taskHeap) Less(i, j int) bool {
	return th[i].at.Unix() < th[j].at.Unix()
}
func (th taskHeap) Swap(i, j int) {
	th[i], th[j] = th[j], th[i]
}

func (th *taskHeap) Push(x interface{}) {
	*th = append(*th, x.(*taskItem))
}

func (th *taskHeap) Pop() interface{} {
	old := *th
	n := len(old)
	x := old[n-1]
	*th = old[0 : n-1]
	return x
}

// HeapTaskPool .
type HeapTaskPool struct {
	sync.WaitGroup
	closed chan bool
	taskOp chan *taskOpInfo
	keys   map[string]*taskItem

	tasks *taskHeap
}

// NewHeapTaskPool .
func NewHeapTaskPool() *HeapTaskPool {
	pool := &HeapTaskPool{}
	pool.Start()
	return pool
}

func (tp *HeapTaskPool) Start() {
	tp.Wait()

	tp.closed = make(chan bool)
	tp.taskOp = make(chan *taskOpInfo, 100)
	tp.keys = make(map[string]*taskItem)

	tp.tasks = &taskHeap{}
	heap.Init(tp.tasks)

	tp.Add(1)
	go tp.loop()
}

func (tp *HeapTaskPool) Stop() {
	close(tp.closed)
	tp.Wait()
}

func execTask(exec TaskFunc, params []interface{}) {
	go exec(params...)
}

func (tp *HeapTaskPool) process() time.Duration {
	for {
		if tp.tasks.Len() <= 0 {
			return 24 * time.Hour
		}
		timeNow := time.Now()
		task := (*tp.tasks)[0]
		if task.cancelled {
			heap.Pop(tp.tasks)
			continue
		}
		if task.at.After(timeNow) {
			return task.at.Sub(timeNow)
		}
		execTask(task.exec, task.params)
		heap.Pop(tp.tasks)
		delete(tp.keys, task.key)
	}
}

func (tp *HeapTaskPool) addOrUpdateTask(t time.Time, key string, exec TaskFunc, params []interface{}) {
	if key == "" {
		key = uuid.NewV4().String()
	} else {
		tp.cancelTask(key)
	}
	taskItem := &taskItem{
		at:     t,
		exec:   exec,
		params: params,
		key:    key,
	}
	tp.keys[key] = taskItem
	heap.Push(tp.tasks, taskItem)
}

func (tp *HeapTaskPool) cancelTask(key string) {
	if task, ok := tp.keys[key]; ok {
		task.cancelled = true
	}
}

func (tp *HeapTaskPool) loop() {
	defer tp.Done()

	var nextTaskInterval time.Duration
	nextTaskInterval = time.Second
	for {
		select {
		case <-tp.closed:
			return
		case taskI := <-tp.taskOp:
			if taskI.opType == taskOpAdd || taskI.opType == taskOpUpdate {
				tp.addOrUpdateTask(taskI.at, taskI.key, taskI.exec, taskI.params)
			} else if taskI.opType == taskOpDel {
				tp.cancelTask(taskI.key)
			}
			nextTaskInterval = tp.process()
		case <-time.After(nextTaskInterval):
			nextTaskInterval = tp.process()
		}
	}
}

func (tp *HeapTaskPool) AddTask(key string, t time.Time, exec TaskFunc, params ...interface{}) error {
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

func (tp *HeapTaskPool) RemoveTask(key string) error {
	if key != "" {
		tp.taskOp <- &taskOpInfo{
			opType: taskOpDel,
			key:    key,
		}
	}
	return nil
}
