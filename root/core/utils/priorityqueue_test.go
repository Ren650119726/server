package utils

import (
	"testing"
)

func TestAglorithm(t *testing.T) {
	heap := NewHeap([]IPriorityInterface{}, 2)
	heap.Push(&MyTimer{})
}
