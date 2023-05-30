package pkg

import "time"

const (
	updateTime = time.Microsecond * 3 // 模拟更新操作的延时
	selectTime = time.Microsecond     // 模拟查询操作的延时
	retryTime  = time.Microsecond * 3 // 遇到死锁重试延迟时间
)
