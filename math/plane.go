package math

import (
	"errors"
)

type Plane struct {
	Norm Vec3
	Dist float64
}

func NewPlane(dot1, dot2, dot3 Vec3) Plane {
	norm := dot2.Sub(dot1).Cross(dot3.Sub(dot1)).Ort()
	dist := -norm.Dot(dot1)
	return Plane{
		norm,
		dist,
	}
}

func (p *Plane) Cross(l Line3) (Vec3, error) {
	d := p.Norm.Dot(l.Vector)
	if d == 0 {
		return Vec3{0, 0, 0}, errors.New("Line is parallel to plane")
	}
	t := -(p.Norm.Dot(l.Point) + p.Dist) / d
	return l.Vector.Mul(t).Add(l.Point), nil
}
