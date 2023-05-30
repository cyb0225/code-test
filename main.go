package main

import (
	"flag"
	"fmt"
	"github.com/cyb0225/code-test/pkg"
	"math/rand"
	"sync"
	"time"
)

const (
	syncWorkers = 10     // 并发执行的 work 数量
	records     = 100000 // 记录数
	works       = 10000  // 每个 worker 工作的数量
	recordMode  = "record"
	tableMode   = "table"
)

var (
	mode string
)

func init() {
	flag.StringVar(&mode, "mode", "record", "隔离模式[table, record]")
	flag.Parse()
}

func main() {
	t := pkg.NewTable()

	var process pkg.Processor
	switch mode {
	case recordMode:
		process = pkg.NewRecordLockProcessor(t)
	case tableMode:
		process = pkg.NewSerializableProcess(t)
	default:
		fmt.Println("")
		return
	}

	// 插入mock数据
	for i := 0; i < records; i++ {
		t.Insert(1)
	}

	rand.NewSource(time.Now().UnixMicro())

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(syncWorkers)
	for i := 0; i < syncWorkers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < works; j++ {
				updateId, selectIds := pkg.GetRandomIds(t.Count(), 3)
				process.Process(updateId, selectIds...)
			}
		}()
	}

	wg.Wait()
	fmt.Println("time usage:", time.Since(start))
}
