package math

import (
	"fmt"
	"math"
)

type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(ov Vec2) Vec2 {
	return Vec2{v.X + ov.X, v.Y + ov.Y}
}

func (v Vec2) Sub(ov Vec2) Vec2 {
	return Vec2{v.X - ov.X, v.Y - ov.Y}
}

func (v Vec2) Dot(ov Vec2) float64 {
	return v.X*ov.X + v.Y*ov.Y
}

func (v Vec2) Mul(n float64) Vec2 {
	return Vec2{v.X * n, v.Y * n}
}

func (v Vec2) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2) Ort() Vec2 {
	l := v.Len()
	if l != 0 {
		return Vec2{v.X / l, v.Y / l}
	}
	return Vec2{0, 0}
}

func (v Vec2) String() string {
	return fmt.Sprintf("(%f, %f)", v.X, v.Y)
}
