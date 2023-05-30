package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Table struct {
	L sync.Mutex
	S []record
}

type record struct {
	data   int
	locker sync.RWMutex
}

func NewTable() *Table {
	return &Table{}
}

// Count
// TODO: 这里有并发风险，但是目前没有并发删除和插入的需求，所以可以忽略不计。
func (t *Table) Count() int {
	return len(t.S)
}

// Insert 插入记录
func (t *Table) Insert(data int) {
	t.L.Lock()
	defer t.L.Unlock()
	t.S = append(t.S, record{data: data})
}

// Select 获取记录
func (t *Table) Select(id int) int {
	time.Sleep(time.Microsecond)
	return t.S[id].data
}

// Update 更新记录
func (t *Table) Update(id int, data int) {
	time.Sleep(time.Microsecond * 3)
	t.S[id].data = data
}

// RLock 对某一行记录加读锁
func (t *Table) RLock(id int) {
	t.S[id].locker.RLock()
}

// Lock 对某一行记录加写锁
func (t *Table) Lock(id int) {
	t.S[id].locker.Lock()
}

// TryLock 对某一行记录加写锁
func (t *Table) TryLock(id int) bool {
	return t.S[id].locker.TryLock()
}

// UnRLock 对某一行记录释放读锁
func (t *Table) UnRLock(id int) {
	t.S[id].locker.RUnlock()
}

// UnLock 对某一行记录释放写锁
func (t *Table) UnLock(id int) {
	t.S[id].locker.Unlock()
}

func (t *Table) ProcessInRecordLockMode() {
	j := rand.Intn(t.Count())
	i := rand.Intn(t.Count())

	pass := false
	for {
		if j == i {
			pass = t.accumulate(j, (i+1)%t.Count(), (i+2)%t.Count())
		} else if j == i+1 {
			pass = t.accumulate(j, i, (i+2)%t.Count())
		} else if j == i+2 {
			pass = t.accumulate(j, i, (i+1)%t.Count())
		} else {
			pass = t.set(j, i, (i+1)%t.Count(), (i+2)%t.Count())
		}

		if pass {
			break
		}

		fmt.Println("retry")
		time.Sleep(time.Millisecond)
	}
}

// 将 select 出来的数据之和累加到 updateId 记录上
func (t *Table) accumulate(updateId int, selectIds ...int) bool {
	// 加读锁可以等待
	sum := 0
	for _, v := range selectIds {
		t.RLock(v)
		sum += t.Select(v)
	}

	defer func() {
		//  释放所有锁
		for _, v := range selectIds {
			t.UnRLock(v)
		}
	}()

	lock := t.TryLock(updateId)
	if !lock {
		// 避免死锁
		fmt.Println("dead lock")
		return false
	}
	defer func() {
		t.UnLock(updateId)
	}()

	sum += t.Select(updateId)
	t.Update(updateId, sum)
	return true
}

// 将 select 出来是数据之和更新到 updateId 记录上
func (t *Table) set(updateId int, selectIds ...int) bool {
	// 加读锁可以等待
	sum := 0
	for _, v := range selectIds {
		t.RLock(v)
		sum += t.Select(v)
	}

	defer func() {
		//  释放所有锁
		for _, v := range selectIds {
			t.UnRLock(v)
		}
	}()

	lock := t.TryLock(updateId)
	if !lock {
		fmt.Println("dead lock")
		// 避免死锁
		return false
	}
	defer func() {
		t.UnLock(updateId)
	}()

	t.Update(updateId, sum)
	return true
}

const (
	syncWorkers = 10     // 并发执行的 work 数量
	records     = 100000 // 记录数
	works       = 10000  // 每个 worker 工作的数量
)

func main() {
	t := NewTable()
	for i := 0; i < records; i++ {
		t.Insert(1)
	}

	rand.NewSource(time.Now().UnixMicro())

	// start set database
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(syncWorkers)
	for i := 0; i < syncWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < works; j++ {
				t.ProcessInRecordLockMode()
			}
		}()
	}

	wg.Wait()
	fmt.Println("time usage:", time.Since(start))
}
