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
	if mode != "record" && mode != "table" {
		fmt.Println("未知的隔离模式:", mode)
		return
	}

	t := pkg.NewTable()
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
				switch mode {
				case recordMode:
					t.ProcessInRecordLockMode()
				case tableMode:
					t.ProcessInSerializableMode()
				}
			}
		}()
	}

	wg.Wait()
	fmt.Println("time usage:", time.Since(start))
}
