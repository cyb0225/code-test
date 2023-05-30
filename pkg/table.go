package pkg

import (
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
	time.Sleep(selectTime)
	return t.S[id].data
}

// Update 更新记录
func (t *Table) Update(id int, data int) {
	time.Sleep(updateTime)
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

// LockTable 加表锁
func (t *Table) LockTable() {
	t.L.Lock()
}

// UnLockTable 释放表锁
func (t *Table) UnLockTable() {
	t.L.Unlock()
}
