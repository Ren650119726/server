// Copyright 2016 The G3N Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package math3d

import (
	"unsafe"
)

// ArrayF32 is a slice of float32 with additional convenience methods
type ArrayF32 []float32

// NewArrayF32 creates a returns a new array of floats
// with the specified initial size and capacity
func NewArrayF32(size, capacity int) ArrayF32 {

	return make([]float32, size, capacity)
}

// Bytes returns the size of the array in bytes
func (a *ArrayF32) Bytes() int {

	return len(*a) * int(unsafe.Sizeof(float32(0)))
}

// Size returns the number of float32 elements in the array
func (a *ArrayF32) Size() int {

	return len(*a)
}

// Len returns the number of float32 elements in the array
// It is equivalent to Size()
func (a *ArrayF32) Len() int {

	return len(*a)
}

// Append appends any number of values to the array
func (a *ArrayF32) Append(v ...float32) {

	*a = append(*a, v...)
}

// AppendVector2 appends any number of Vector2 to the array
func (a *ArrayF32) AppendVector2(v ...*Vector2) {

	for i := 0; i < len(v); i++ {
		*a = append(*a, v[i].X, v[i].Y)
	}
}

// AppendVector3 appends any number of Vector3 to the array
func (a *ArrayF32) AppendVector3(v ...*Vector3) {

	for i := 0; i < len(v); i++ {
		*a = append(*a, v[i].X, v[i].Y, v[i].Z)
	}
}

// AppendColor appends any number of Color to the array
func (a *ArrayF32) AppendColor(v ...*Color) {

	for i := 0; i < len(v); i++ {
		*a = append(*a, v[i].R, v[i].G, v[i].B)
	}
}

// AppendColor4 appends any number of Color4 to the array
func (a *ArrayF32) AppendColor4(v ...*Color4) {

	for i := 0; i < len(v); i++ {
		*a = append(*a, v[i].R, v[i].G, v[i].B, v[i].A)
	}
}

// GetVector2 stores in the specified Vector2 the
// values from the array starting at the specified pos.
func (a ArrayF32) GetVector2(pos int, v *Vector2) {

	v.X = a[pos]
	v.Y = a[pos+1]
}

// GetVector3 stores in the specified Vector3 the
// values from the array starting at the specified pos.
func (a ArrayF32) GetVector3(pos int, v *Vector3) {

	v.X = a[pos]
	v.Y = a[pos+1]
	v.Z = a[pos+2]
}

// GetColor stores in the specified Color the
// values from the array starting at the specified pos
func (a ArrayF32) GetColor(pos int, v *Color) {

	v.R = a[pos]
	v.G = a[pos+1]
	v.B = a[pos+2]
}

// Set sets the values of the array starting at the specified pos
// from the specified values
func (a ArrayF32) Set(pos int, v ...float32) {

	for i := 0; i < len(v); i++ {
		a[pos+i] = v[i]
	}
}

// SetVector2 sets the values of the array at the specified pos
// from the XY values of the specified Vector2
func (a ArrayF32) SetVector2(pos int, v *Vector2) {

	a[pos] = v.X
	a[pos+1] = v.Y
}

// SetVector3 sets the values of the array at the specified pos
// from the XYZ values of the specified Vector3
func (a ArrayF32) SetVector3(pos int, v *Vector3) {

	a[pos] = v.X
	a[pos+1] = v.Y
	a[pos+2] = v.Z
}

// SetColor sets the values of the array at the specified pos
// from the RGB values of the specified Color
func (a ArrayF32) SetColor(pos int, v *Color) {

	a[pos] = v.R
	a[pos+1] = v.G
	a[pos+2] = v.B
}

// SetColor4 sets the values of the array at the specified pos
// from the RGBA values of specified Color4
func (a ArrayF32) SetColor4(pos int, v *Color4) {

	a[pos] = v.R
	a[pos+1] = v.G
	a[pos+2] = v.B
	a[pos+3] = v.A
}

// ArrayU32 is a slice of uint32 with additional convenience methods
type ArrayU32 []uint32

// NewArrayU32 creates a returns a new array of uint32
// with the specified initial size and capacity
func NewArrayU32(size, capacity int) ArrayU32 {

	return make([]uint32, size, capacity)
}

// Bytes returns the size of the array in bytes
func (a *ArrayU32) Bytes() int {

	return len(*a) * int(unsafe.Sizeof(uint32(0)))
}

// Size returns the number of float32 elements in the array
func (a *ArrayU32) Size() int {

	return len(*a)
}

// Len returns the number of float32 elements in the array
func (a *ArrayU32) Len() int {

	return len(*a)
}

// Append appends n elements to the array updating the slice if necessary
func (a *ArrayU32) Append(v ...uint32) {

	*a = append(*a, v...)
}
