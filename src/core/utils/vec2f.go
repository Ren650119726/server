package utils

import (
	"fmt"
	"math"
)

type Vec2f struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
}

func Add2d(l, r Vec2f) Vec2f {
	return Vec2f{
		X: l.X + r.X,
		Y: l.Y + r.Y,
	}
}

func Sub2d(l, r Vec2f) Vec2f {
	return Vec2f{
		X: l.X - r.X,
		Y: l.Y - r.Y,
	}
}

func Mul2d(v Vec2f, d float32) Vec2f {
	return Vec2f{
		X: v.X * d,
		Y: v.Y * d,
	}
}

func Div2d(v Vec2f, d float32) Vec2f {
	return Vec2f{
		X: v.X / d,
		Y: v.Y / d,
	}
}

func Dot2d(l, r Vec2f) float32 {
	return l.X*r.X + l.Y*r.Y
}

func (v Vec2f) SqrMagnitude() float32 {
	return v.X*v.X + v.Y*v.Y
}

// 向量模长
func (v Vec2f) Magnitude() float32 {
	return float32(math.Sqrt(float64(v.SqrMagnitude())))
}

// 计算单位向量
func Normalize2d(v Vec2f) Vec2f {
	d := v.X*v.X + v.Y*v.Y
	if IsEqualZero32(d) {
		return Vec2f{0, 0}
	}

	d = float32(math.Sqrt(float64(d)))
	return Mul2d(v, 1/d)
}

func (v Vec2f) String() string {
	return fmt.Sprintf("Vec2f(%.3f, %.3f)", v.X, v.Y)
}
