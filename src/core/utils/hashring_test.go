package utils

import (
	"fmt"
	"testing"
)

func TestHashRing_GetNode(t *testing.T) {
	hashring := NewHashRing(0)
	hashring.AddNode("1", 1)
	hashring.AddNode("2", 1)
	hashring.AddNode("3", 1)
	for i := 0; i < 10; i++ {
		k := fmt.Sprintf("%d", i)
		fmt.Println("******** key = ", k, " ---- node = ", hashring.GetNode(k))
	}
}

func BenchmarkHashRing_GetNode(b *testing.B) {
	hashring := NewHashRing(0)
	for i := 0; i < 70; i++ {
		k := fmt.Sprintf("%d", i)
		hashring.AddNode(k, 1)
	}

	for i := 0; i < b.N; i++ {
		k := fmt.Sprintf("asfsdfdsfsxvxcvcxfdfs%d", i)
		hashring.GetNode(k)
	}
}

/**
效率:
10个节点		627 ns/op
60个节点		1158 ns/op
70个节点		1296 ns/op
80个节点		2263 ns/op
90个节点		1025236500 ns/op	1.0252365秒(s)
100个节点		1254853000 ns/op	1.254853秒(s)
*/
