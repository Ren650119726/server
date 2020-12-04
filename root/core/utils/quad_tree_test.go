package utils

import (
	"runtime"
	"testing"
)

func TestQuadTree(t *testing.T) {
	memStats := &runtime.MemStats{}
	runtime.ReadMemStats(memStats)
	tree := NewQuadTree(4,Vec2f{0,2000},Vec2f{2000,0})
	runtime.ReadMemStats(memStats)
	println("%v",tree)
}
