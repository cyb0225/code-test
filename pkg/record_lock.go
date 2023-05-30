package pkg

import (
	"fmt"
	"math/rand"
	"time"
)

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
		time.Sleep(retryTime)
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
