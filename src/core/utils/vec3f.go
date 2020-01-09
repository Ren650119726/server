package utils

import (
	"fmt"
	"math"
)

type Vec3f struct {
	X float32 `json:"x"`
	Y float32 `json:"y"`
	Z float32 `json:"z"`
}

func Add(l, r Vec3f) Vec3f {
	return Vec3f{
		X: l.X + r.X,
		Y: l.Y + r.Y,
		Z: l.Z + r.Z,
	}
}

func Sub(l, r Vec3f) Vec3f {
	return Vec3f{
		X: l.X - r.X,
		Y: l.Y - r.Y,
		Z: l.Z - r.Z,
	}
}

func Mul(v Vec3f, d float32) Vec3f {
	return Vec3f{
		X: v.X * d,
		Y: v.Y * d,
		Z: v.Z * d,
	}
}

func Div(v Vec3f, d float32) Vec3f {
	return Vec3f{
		X: v.X / d,
		Y: v.Y / d,
		Z: v.Z / d,
	}
}

func Dot(l, r Vec3f) float32 {
	return l.X*r.X + l.Y*r.Y + l.Z*r.Z
}

func (v Vec3f) SqrMagnitude() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// 向量模长
func (v Vec3f) Magnitude() float32 {
	return float32(math.Sqrt(float64(v.SqrMagnitude())))
}

// 计算单位向量
func Normalize(v Vec3f) Vec3f {
	d := v.X*v.X + v.Y*v.Y + v.Z*v.Z
	if IsEqualZero32(d) {
		return Vec3f{0, 0, 0}
	}

	d = float32(math.Sqrt(float64(d)))
	return Mul(v, 1/d)
}

func (v Vec3f) String() string {
	return fmt.Sprintf("Vec3f(%.3f, %.3f, %.3f)", v.X, v.Y, v.Z)
}
