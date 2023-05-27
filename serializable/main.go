package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Table struct {
	S []record
	L sync.Mutex
}

type record struct {
	data int
}

func NewTable() *Table {
	return &Table{}
}

func (t *Table) Insert(data int) {
	t.L.Lock()
	defer t.L.Unlock()
	insertRecord := record{data: data}
	t.S = append(t.S, insertRecord)
}

func (t *Table) Select(id int) int {
	time.Sleep(time.Microsecond)
	return t.S[id].data
}

func (t *Table) Update(id int, data int) {
	time.Sleep(time.Microsecond * 3)
	t.S[id].data = data
}

func (t *Table) StartTransaction() {
	t.L.Lock()
}

func (t *Table) Commit() {
	t.L.Unlock()
}

const (
	syncWorkers = 10     // 并发执行的 work 数量
	records     = 100000 // 记录数
	works       = 10000  // 每个 worker 工作的数量
)

func (t *Table) process() {
	j := rand.Intn(records)
	i := rand.Intn(records)
	t.StartTransaction()
	t.Update(j, t.Select(i%records)+t.Select((i+1)%records)+t.Select((i+2)%records))
	t.Commit()
}

func main() {
	// init table and insert records
	t := NewTable()
	for i := 0; i < records; i++ {
		t.Insert(1)
	}

	rand.NewSource(time.Now().UnixMicro())

	// start update database
	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(syncWorkers)
	for i := 0; i < syncWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < works; j++ {
				t.process()
			}
		}()
	}

	wg.Wait()
	fmt.Println("time usage:", time.Since(start))
}
