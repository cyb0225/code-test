package pkg

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewRecordLockProcessor(t *testing.T) {
	p := NewRecordLockProcessor(NewTable())
	assert.NotNil(t, p)
}

func TestRecordLockProcessor_Process(t1 *testing.T) {
	t1.Run("测试存在相同的id", func(t1 *testing.T) {
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &SerializableProcess{t: t}
		s.Process(0, 0, 1)
		assert.Equal(t1, 2, t.S[0].data)
	})

	t1.Run("测试不存在相同id", func(t1 *testing.T) {
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &SerializableProcess{t: t}
		s.Process(0, 1, 2)
		assert.Equal(t1, 2, t.S[0].data)
	})

	t1.Run("测试多个程序的并发正确性，存在相同id", func(t1 *testing.T) {
		count := 16 // 并发协程
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &SerializableProcess{t: t}
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				defer wg.Done()
				s.Process(0, 0, 1)
			}()
		}

		wg.Wait()
		assert.Equal(t1, 17, t.S[0].data)
	})

	t1.Run("测试多个程序的并发正确性，不存在相同id", func(t1 *testing.T) {
		count := 16 // 并发协程
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &RecordLockProcessor{t: t}
		var wg sync.WaitGroup
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				defer wg.Done()
				s.Process(0, 1, 2)
			}()
		}

		wg.Wait()
		assert.Equal(t1, 2, t.S[0].data)
	})

}

func TestRecordLockProcessor_accumulate(t1 *testing.T) {
	t1.Run("测试程序正确性", func(t1 *testing.T) {
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &RecordLockProcessor{t: t}
		s.accumulate(0, 1, 2)
		assert.Equal(t1, 3, t.S[0].data)
	})

	// 多线程情况下会有 tryLock 这个不确定性因素，所以统一在 process 的顶层模块进行测试
}

func TestRecordLockProcessor_set(t1 *testing.T) {
	t1.Run("测试程序正确性", func(t1 *testing.T) {
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &RecordLockProcessor{t: t}
		s.Process(0, 1, 2)
		assert.Equal(t1, 2, t.S[0].data)
	})

	// 多线程情况下会有 tryLock 这个不确定性因素，所以统一在 process 的顶层模块进行测试
}
