# schedule_task
`golang`任务延迟执行实现

```
func testTaskPool(t *testing.T, pool ScheduleTaskPool) {
	err := pool.AddTask("10", time.Now().Add(5*time.Second), func(args ...interface{}) {
		v := args[0].([]int)
		t.Log(v)
	}, []int{1, 2})
	assert.Nil(t, err)
	err = pool.AddTask("", time.Now().Add(2*time.Second), func(args ...interface{}) {
		v := args[0].([]int)
		t.Log(v)
	}, []int{3})
	assert.Nil(t, err)
	err = pool.AddTask("", time.Now().Add(2*time.Second), func(args ...interface{}) {
		v := args[0].([]int)
		t.Log(v)
	}, []int{4})
	assert.Nil(t, err)
	err = pool.AddTask("10", time.Now().Add(9*time.Second), func(args ...interface{}) {
		v := args[0].([]int)
		t.Log(v)
	}, []int{5, 6, 7, 8, 9})
	assert.Nil(t, err)
}

func TestTimeWheelTaskPool(t *testing.T) {
	testTaskPool(t, CreateTimeWheelTaskPool())
	testTaskPool(t, CreateHeapTaskPool())

	time.Sleep(10 * time.Second)
}

```
