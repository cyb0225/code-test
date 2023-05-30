package pkg

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixMicro())
}

type Processor interface {
	// Process 实现处理逻辑，针对表，实现将 selectId 的数据和更新到 updateId 中
	Process(updateId int, selectIds ...int)
}

// GetRandomIds 获取随机的更新对象,以及查询对象
// 传入 range，即 id 的范围 [0, rg)
// 传入的 n 代表要返回的 selectId 的数量
func GetRandomIds(rg int, n int) (updateId int, selectIds []int) {
	updateId = rand.Intn(rg)
	selectIds = make([]int, n)
	selectIds[0] = rand.Intn(rg)
	for i := 1; i < len(selectIds); i++ {
		selectIds[i] = (selectIds[0] + i) % rg
	}

	return
}
