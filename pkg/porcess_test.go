package pkg

import (
	"testing"
)

func TestGetRandomIds(t *testing.T) {
	const (
		rg = 10
		n  = 3
	)

	//  通过大量的 for 循环测试是否会产生边界问题
	for i := 0; i < 100; i++ {
		ids, selectIds := GetRandomIds(rg, n)
		if len(selectIds) != n {
			t.Fatalf("the length of selectIds is not %d\n", n)
		}

		if ids < 0 || ids >= 10 {
			t.Fatalf("update id %d is invalid\n", ids)
		}

		for _, v := range selectIds {
			if v < 0 || v >= 10 {
				t.Fatalf("select id %d is invalid\n", v)
			}
		}
	}
}
