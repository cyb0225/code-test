package pkg

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestNewTable(t *testing.T) {
	table := NewTable()
	assert.NotNil(t, table)
}

func TestTable_Count(t1 *testing.T) {
	t := &Table{}
	count := 10
	for i := 0; i < count; i++ {
		t.S = append(t.S, record{data: i})
	}
	assert.Equal(t1, count, t.Count())
}

func TestTable_Insert(t1 *testing.T) {
	t := &Table{}
	data := 10
	t.Insert(data)
	assert.Equal(t1, data, t.S[0].data)
}

func TestTable_Lock(t1 *testing.T) {
	t := &Table{}
	t.S = make([]record, 10)
	id := 0
	data := 10
	t.S[id] = record{data: data}
	t.Lock(id)
	ch := make(chan struct{})
	go func() {
		ch <- struct{}{}
		t.Lock(id)
		t1.Error("lock record failed")
	}()

	<-ch
	time.Sleep(time.Microsecond * 10)
}

func TestTable_LockTable(t1 *testing.T) {
	t := &Table{}
	t.LockTable()
	waitTime := time.Second
	ch1 := make(chan struct{})
	go func() {
		ch1 <- struct{}{}
		start := time.Now()
		t.LockTable()
		assert.True(t1, time.Since(start) > waitTime) // 代表被锁住了，同时能执行下一步说明已经解锁了
	}()

	<-ch1
	t.UnLockTable()
}

func TestTable_RLock(t1 *testing.T) {
	t := &Table{}
	t.S = make([]record, 10)
	id := 0
	data := 10
	t.S[id] = record{data: data}

	// 3个读锁获取数据，一个写锁写数据
	// 当读协程启动并加读锁后的时候，判断写协程是否会正常阻塞
	// 写协程启动后，再使用 cond 通知所有读协程释放锁后加锁
	readCount := 3
	num := 0

	var wg sync.WaitGroup
	c := sync.NewCond(&sync.Mutex{})

	wg.Add(readCount)
	for i := 0; i < readCount; i++ {
		gid := i
		go func() {
			defer wg.Done()
			t1.Log(gid, "读协程启动...")
			c.L.Lock()
			defer c.L.Unlock()

			// 添加锁
			t.RLock(id)
			defer t.UnRLock(id)

			num++
			c.Wait() // 等待写进程
		}()
	}

	// 写协程
	wg.Add(1)
	go func() {
		defer wg.Done()
		t1.Log("写协程启动...")

		start := time.Now()
		sleepTime := time.Second
		// 异步调用
		go func() {
			time.Sleep(sleepTime)
			c.Broadcast()
		}()

		t.Lock(id)
		defer t.UnLock(id)
		assert.True(t1, time.Since(start) < sleepTime) // 代表没有锁住
	}()

	wg.Wait()
	assert.Equal(t1, readCount, num)
}

func TestTable_Select(t1 *testing.T) {
	t := &Table{}
	t.S = make([]record, 10)
	id := 0
	data := 10
	t.S[id] = record{data: data}
	assert.Equal(t1, data, t.Select(id))
}

func TestTable_TryLock(t1 *testing.T) {
	t := &Table{}
	t.S = make([]record, 10)
	id := 0
	data := 10
	t.S[id] = record{data: data}

	// 先加锁,然后开启协程判断是否能加锁，如果返回 false 则表示trylock无法获取锁
	t.Lock(id)
	ch1 := make(chan struct{})
	ch2 := make(chan struct{})
	go func() {
		assert.False(t1, t.TryLock(id)) // 加不上锁
		ch1 <- struct{}{}               // 通知协程可以解锁
		<-ch2                           // 其他协程表示已经解锁完毕
		assert.True(t1, t.TryLock(id))
		ch1 <- struct{}{} // 告诉主协程可以退出了
	}()

	<-ch1
	t.UnLock(id) // 解锁
	ch2 <- struct{}{}
	<-ch1
}

func TestTable_Update(t1 *testing.T) {
	t := &Table{}
	t.S = make([]record, 10)
	id := 0
	data := 10
	t.S[id] = record{data: data}

	updateData := 12
	t.Update(id, updateData)
	assert.Equal(t1, updateData, t.S[id].data)
}
