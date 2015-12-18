package math

import (
	"encoding/json"
	"fmt"
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

func (v Vec3) Add(ov Vec3) Vec3 {
	return Vec3{v.X + ov.X, v.Y + ov.Y, v.Z + ov.Z}
}

func (v Vec3) Sub(ov Vec3) Vec3 {
	return Vec3{v.X - ov.X, v.Y - ov.Y, v.Z - ov.Z}
}

func (v Vec3) Len() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

func (v Vec3) Dot(ov Vec3) float64 {
	return v.X*ov.X + v.Y*ov.Y + v.Z*ov.Z
}

func (v Vec3) Cross(ov Vec3) Vec3 {
	return Vec3{
		v.Y*ov.Z - v.Z*ov.Y,
		v.Z*ov.X - v.X*ov.Z,
		v.X*ov.Y - v.Y*ov.X,
	}
}

func (v Vec3) Mul(n float64) Vec3 {
	return Vec3{v.X * n, v.Y * n, v.Z * n}
}

func (v Vec3) Ort() Vec3 {
	l := v.Len()
	if l != 0 {
		return Vec3{v.X / l, v.Y / l, v.Z / l}
	} else {
		return Vec3{0, 0, 0}
	}
}

func (v *Vec3) UnmarshalJSON(data []byte) error {
	var a [3]float64
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	v.X, v.Y, v.Z = a[0], a[1], a[2]
	return nil
}
func (v *Vec3) MarshalJSON() ([]byte, error) {
	a := [3]float64{v.X, v.Y, v.Z}
	if data, err := json.Marshal(a); err != nil {
		return nil, err
	} else {
		return data, nil
	}
}

func (v Vec3) String() string {
	return fmt.Sprintf("(%f, %f, %f)", v.X, v.Y, v.Z)
}
