package main

import (
	"sync"
	"testing"
)

func TestGo(t *testing.T) {
	m := make(map[int]int,0)
	for i := 0;i < 10000;i++{
		m[i] = i*100
	}
	w := sync.WaitGroup{}
	w.Add(1)
	go func() {
		println("go1")
		for i:=10000;i < 1000000;i++{
			m[i] = i*100
		}
		w.Done()
	}()

	w.Add(1)
	go func() {
		println("go2")
		for k,v := range m{
			println(2,k,v)
			//m[k] = 123
		}
		w.Done()
	}()

	w.Add(1)
	go func() {
		println("go3")
		for k,v := range m{
			println(3,k,v)
			//m[k] = 456
		}
		w.Done()
	}()
	w.Wait()
}
