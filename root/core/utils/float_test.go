package utils

import (
	"fmt"
	"testing"
)

func TestQuadratic(t *testing.T) {
	x1, x2, _ := Quadratic(1, 2, -3)
	//fmt.Println("result:", d)
	fmt.Println(x1, x2)
	//if d != 1 {
	//	t.Fail()
	//}
}

func BenchmarkQuadratic(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Quadratic(-11, -100, 125)
	}
}
func TestMaxf32(t *testing.T) {
	println(Maxf32(1.2, 33.2, 333, 34.0, 0.1, 1.6))
}

func TestFloatCut(t *testing.T) {
	f := 1.2314612312
	n := 4
	value := FloatCut(f, uint32(n))
	fmt.Println(value)
}

func BenchmarkFloatCut(b *testing.B) {
	f := 1.2314312312
	n := 4
	for i := 0; i < b.N; i++ {
		FloatCut(f, uint32(n))
	}
}
