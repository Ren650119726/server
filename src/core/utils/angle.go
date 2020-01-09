package utils

import (
	"math"
)

const (
	PI      = math.Pi
	Deg2Rad = PI / 180
	Rad2Deg = 180 / PI

	ANGLE_22_5 = 22.5
	ANGLE_45   = 45
	ANGLE_90   = 90
	ANGLE_120  = 120
	ANGLE_180  = 180
)

var (
	CosineOneDeg = float32(math.Cos(1 * Deg2Rad))
)

// 给定一个角度, 返回一个0-360表示的的角度
func Angle360(angle float64) float64 {
	// 当angle大于75000时，把angle规范到0-360,使用mod函数的效率更高
	if angle < -75000 || angle > 75000 {
		return math.Mod(angle, 360)
	}

	// 把负数角度修正为整数角度
	for angle < 0 {
		angle += 360
	}

	// 把大于360度的角度修正为360度的角度
	for angle >= 360 {
		angle -= 360
	}

	return angle
}

// 两个角度之间的夹角
func DiffAngleAbs(src, dst float64) float64 {
	diff := Angle360(src) - Angle360(dst)
	if diff < -180 {
		diff += 360
	} else if diff > 180 {
		diff -= 360
	}
	return math.Abs(diff)
}

// 角度差[-180, 180], 逆时针为正, 顺时针为负
func DiffAngle180(src, dst float64) float64 {
	diff := dst - src
	diff = Angle360(diff)

	if diff > 180 {
		diff = diff - 360
	} else if diff < -180 {
		diff = diff + 360

	}
	return diff
}

// 判断目标角度是否在源角度的顺时针方向
func IsClockwise(src, dst float64) bool {
	src, dst = Angle360(src), Angle360(dst)
	if dst < src {
		dst += 360
	}
	dif := dst - src
	return dif > 180
}

func HorizonAngle(srcX float32, srcZ float32, dstX float32, dstZ float32) float64 {
	dx := dstX - srcX
	dz := dstZ - srcZ

	rad := math.Atan2(float64(dz), float64(dx))
	angle := rad * 360 / (2 * math.Pi)
	return Angle360(angle)
}

// 将角度转换为弧度
func Angle2Radian(angle float64) float64 {
	return angle * math.Pi / 180.0
}

func Radian2Angle(radian float64) float64 {
	return radian * 180.0 / math.Pi
}

// 角度是否合法
func IsLegalAngle(angle float64) bool {
	if angle < 0 || angle > 360 || math.IsNaN(angle) {
		return false
	}
	return true
}

// 判断一个角是否在两个角之间
func IsAngleBetween(angle, left, right float64) bool {
	diff1 := DiffAngleAbs(left, right)
	return DiffAngleAbs(angle, left) < diff1 && DiffAngleAbs(angle, right) < diff1
}
