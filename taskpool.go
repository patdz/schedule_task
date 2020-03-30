package schedule_task

import "time"

// TaskFunc .
type TaskFunc func(args ...interface{})

type ScheduleTaskPool interface {
	AddTask(key string, t time.Time, exec TaskFunc, params ...interface{}) error
	RemoveTask(key string) error
}

// CreateTimeWheelTaskPool ...
func CreateTimeWheelTaskPool() ScheduleTaskPool {
	return NewTimeWheelTaskPool()
}

// CreateHeapTaskPool .
func CreateHeapTaskPool() ScheduleTaskPool {
	return NewHeapTaskPool()
}
