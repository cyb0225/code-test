package pkg

import (
	"fmt"
	"time"
)

const (
	retryTime = time.Microsecond * 3 // 遇到死锁重试延迟时间
)

var _ Processor = (*RecordLockProcessor)(nil)

type RecordLockProcessor struct {
	t *Table
}

func NewRecordLockProcessor(t *Table) *RecordLockProcessor {
	return &RecordLockProcessor{
		t: t,
	}
}

func (r *RecordLockProcessor) Process(updateId int, selectIds ...int) {
	// 判断是否有 updateId 与 selectId 重复的情况，selectId 各个字段要保证不会重复
	hasSame := false
	for i, v := range selectIds {
		if updateId == v {
			hasSame = true
			if i == len(selectIds)-1 {
				selectIds = selectIds[:len(selectIds)-1]
			} else {
				selectIds = append(selectIds[:i], selectIds[i+1:]...)
			}
			break
		}
	}

	pass := false // 判断是否遇到死锁需要重新执行
	for {
		if hasSame {
			pass = r.accumulate(updateId, selectIds...)
		} else {
			pass = r.set(updateId, selectIds...)
		}

		// 程序正常执行
		if pass {
			return
		}

		// 遇到死锁, 睡眠等待其他事务执行完毕后再次执行
		fmt.Println("retry")
		time.Sleep(retryTime)
	}
}

// 将 select 出来的数据之和累加到 updateId 记录上
func (r *RecordLockProcessor) accumulate(updateId int, selectIds ...int) bool {
	// 加读锁可以等待
	sum := 0
	for _, v := range selectIds {
		r.t.RLock(v)
		sum += r.t.Select(v)
	}

	defer func() {
		//  释放所有锁
		for _, v := range selectIds {
			r.t.UnRLock(v)
		}
	}()

	lock := r.t.TryLock(updateId)
	if !lock {
		// 避免死锁
		fmt.Println("dead lock")
		return false
	}
	defer func() {
		r.t.UnLock(updateId)
	}()

	sum += r.t.Select(updateId)
	r.t.Update(updateId, sum)
	return true
}

// 将 select 出来是数据之和更新到 updateId 记录上
func (r *RecordLockProcessor) set(updateId int, selectIds ...int) bool {
	// 加读锁可以等待
	sum := 0
	for _, v := range selectIds {
		r.t.RLock(v)
		sum += r.t.Select(v)
	}

	defer func() {
		//  释放所有锁
		for _, v := range selectIds {
			r.t.UnRLock(v)
		}
	}()

	lock := r.t.TryLock(updateId)
	if !lock {
		fmt.Println("dead lock")
		// 避免死锁
		return false
	}
	defer func() {
		r.t.UnLock(updateId)
	}()

	r.t.Update(updateId, sum)
	return true
}
