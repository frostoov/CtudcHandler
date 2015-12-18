package math

import (
	"errors"
)

type Quadrangle3 struct {
	Vertices [4]Vec3
	coord    CoordSystem
	plane    Plane
	quad2    Quadrangle2
}

func NewQuadrangle3(v []Vec3) Quadrangle3 {
	coord := getCoordSystem(v)
	return Quadrangle3{
		Vertices: [4]Vec3{v[0], v[1], v[2], v[3]},
		coord:    coord,
		plane:    NewPlane(v[0], v[1], v[2]),
		quad2:    getQuadrangle2(v, coord),
	}
}

func (q *Quadrangle3) HasPoint(p Vec3) bool {
	t := q.coord.ConvertVector(p)
	return q.quad2.HasPoint(Vec2{t.X, t.Y})
}

func (q *Quadrangle3) Cross(l Line3) (Vec3, error) {
	c, err := q.plane.Cross(l)
	if err != nil || !q.HasPoint(c) {
		return Vec3{0, 0, 0}, errors.New("(q Quadrangle3) Cross")
	}
	return c, nil
}

func getCoordSystem(v []Vec3) CoordSystem {
	ox := v[1].Sub(v[0]).Ort()
	oy := v[3].Sub(v[0]).Ort()
	oz := ox.Cross(oy).Ort()
	return CoordSystem{v[0], ox, oy, oz}
}

func getQuadrangle2(v3 []Vec3, coord CoordSystem) Quadrangle2 {
	var v2 [4]Vec2
	for i := 0; i < len(v2); i++ {
		t := coord.ConvertVector(v3[i])
		v2[i] = Vec2{t.X, t.Y}
	}
	return NewQuadrangle2(v2[:])
}
