package schedule_task

import (
	"container/heap"
	"testing"
	"time"
)

func Test_taskHeap(t *testing.T) {
	h := &taskHeap{}
	heap.Init(h)

	heap.Push(h, &taskItem{
		at: time.Now().Add(10 * time.Second),
	})
	heap.Push(h, &taskItem{
		at: time.Now().Add(8 * time.Second),
	})
	heap.Push(h, &taskItem{
		at: time.Now().Add(12 * time.Second),
	})
	v := (*h)[0]
	t.Log(v.at)
	vi := heap.Pop(h)
	t.Log(vi.(*taskItem).at)

	v = (*h)[0]
	t.Log(v.at)
	vi = heap.Pop(h)
	t.Log(vi.(*taskItem).at)

	v = (*h)[0]
	t.Log(v.at)
	vi = heap.Pop(h)
	t.Log(vi.(*taskItem).at)
}
