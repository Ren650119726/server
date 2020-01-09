package utils

import (
	"errors"
	"fmt"
	"math"
)

const Epsilon = 0.000001
const SqrEpsilon = Epsilon * Epsilon

// 精度使用可选参数
func IsEqual(f1, f2 float64, epsilon ...float64) bool {
	if len(epsilon) > 0 {
		return math.Abs(f1-f2) < epsilon[0]
	} else {
		return math.Abs(f1-f2) < Epsilon
	}
}

func IsEqual32(f1, f2 float32, epsilon ...float32) bool {
	if len(epsilon) > 0 {
		return IsEqual(float64(f1), float64(f2), float64(epsilon[0]))
	} else {
		return IsEqual(float64(f1), float64(f2))
	}
}

func IsEqualZero(f float64) bool {
	if f < 0 {
		return f > -Epsilon
	} else {
		return f < Epsilon
	}
}

func IsEqualZero32(f float32) bool {
	return IsEqual32(f, 0)
}

func IsLower(f1, f2 float64, epsilon ...float64) bool {
	if len(epsilon) > 0 {
		return f1-f2 <= -epsilon[0]
	} else {
		return f1-f2 <= -Epsilon
	}
}

func IsLower32(f1, f2 float32, epsilon ...float32) bool {
	if len(epsilon) > 0 {
		return IsLower(float64(f1), float64(f2), float64(epsilon[0]))
	} else {
		return IsLower(float64(f1), float64(f2))
	}
}

func IsLowerEqual(f1, f2 float64, epsilon ...float64) bool {
	return IsLower(f1, f2, epsilon...) || IsEqual(f1, f2, epsilon...)
}

func IsLowerEqual32(f1, f2 float32, epsilon ...float32) bool {
	if len(epsilon) > 0 {
		return IsLowerEqual(float64(f1), float64(f2), float64(epsilon[0]))
	} else {
		return IsLowerEqual(float64(f1), float64(f2))
	}
}

func IsGreater(f1, f2 float64, epsilon ...float64) bool {
	if len(epsilon) > 0 {
		return f1-f2 >= epsilon[0]
	} else {
		return f1-f2 >= Epsilon
	}
}

func IsGreater32(f1, f2 float32, epsilon ...float32) bool {
	if len(epsilon) > 0 {
		return IsGreater(float64(f1), float64(f2), float64(epsilon[0]))
	} else {
		return IsGreater(float64(f1), float64(f2))
	}
}

func IsGreaterEqual(f1, f2 float64, epsilon ...float64) bool {
	return IsGreater(f1, f2, epsilon...) || IsEqual(f1, f2, epsilon...)
}

func IsGreaterEqual32(f1, f2 float32, epsilon ...float32) bool {
	if len(epsilon) > 0 {
		return IsGreaterEqual(float64(f1), float64(f2), float64(epsilon[0]))
	} else {
		return IsGreaterEqual(float64(f1), float64(f2))
	}
}

func IsNear(f1, f2, v float64) bool {
	return math.Abs(f1-f2) < math.Abs(v)
}

func Round(v float64) float64 {
	return math.Floor(v + .5)
}

func RoundToInt(v float64) int {
	return int(math.Floor(v + .5))
}

func RoundToInt64(v float64) int64 {
	return int64(math.Floor(v + .5))
}

func Clamp(v, min, max float64) float64 {
	if v < min {
		return min
	} else if v > max {
		return max
	}

	return v
}

func Clamp32(v, min, max float32) float32 {
	if v < min {
		return min
	} else if v > max {
		return max
	}

	return v
}

//ax2+bx+c=0,求解x1,x2
func Quadratic(a, b, c float64) (x1, x2 float64, err error) {
	if a == 0 {
		panic(fmt.Sprintf("求解二元一次方程参数异常, a=%v, b=%v, c=%v", a, b, c))
	}
	p := b*b - 4*a*c
	if p < 0 {
		err = errors.New(fmt.Sprintf("方程式无解,p=%v", p))
		return
	}
	p = math.Sqrt(p)
	x1 = (-b + p) / (2 * a)
	x2 = (-b - p) / (2 * a)
	return x1, x2, nil
}

// 返回一系列float32中的最大值
func Maxf32(v1 float32, vs ...float32) float32 {
	if len(vs) < 1 {
		return v1
	}

	var max = v1
	for _, v := range vs {
		if max < v {
			max = v
		}
	}

	return max
}

// 返回一系列int中的最大值
func MaxInt(v1 int, vs ...int) int {
	if len(vs) < 1 {
		return v1
	}

	var max = v1
	for _, v := range vs {
		if max < v {
			max = v
		}
	}
	return max
}

// 返回一系列int中的最小值
func Min(v1 int, vs ...int) int {
	if len(vs) < 1 {
		return v1
	}

	var min = v1
	for _, v := range vs {
		if v < min {
			min = v
		}
	}
	return min
}

// 在X/Z两个方向上的分量
func HorizonComp(distance float32, angle float64) (x, z float32) {
	distanceX := distance * float32(math.Cos(Angle2Radian(angle)))
	distanceZ := distance * float32(math.Sin(Angle2Radian(angle)))
	return distanceX, distanceZ
}

func Abs32(v float32) float32 {
	if v < 0 {
		v = -v
	}
	return v
}

func Abs(v int) int {
	if v < 0 {
		v = -v
	}
	return v
}

//保留浮点数后n位
func FloatCut(v float64, n uint32) float64 {
	pow10_n := math.Pow10(int(n))
	negative_pow10_n := math.Pow10(-int(n))
	value := math.Trunc(v*pow10_n+0.5) * negative_pow10_n
	return value
}

func Pow32(x, y float32) float32 {
	return float32(math.Pow(float64(x), float64(y)))
}
