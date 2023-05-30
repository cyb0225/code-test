package pkg

import (
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestNewSerializableProcess(t *testing.T) {
	s := NewSerializableProcess(NewTable())
	assert.NotNil(t, s)
}

func TestSerializableProcess_Process(t1 *testing.T) {
	t1.Run("测试程序正确性", func(t1 *testing.T) {
		t := NewTable()
		t.S = make([]record, 3)
		for i := 0; i < 3; i++ {
			t.S[i] = record{data: 1}
		}

		s := &SerializableProcess{t: t}
		s.Process(0, 1, 2)
		assert.Equal(t1, 2, t.S[0].data)
	})

	t1.Run("测试多个程序的并发正确性", func(t1 *testing.T) {
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
				s.Process(0, 1, 2)
			}()
		}

		wg.Wait()
		assert.Equal(t1, 2, t.S[0].data)
	})
}
