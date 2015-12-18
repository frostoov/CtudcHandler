package math

type Quadrangle2 struct {
	Vertices [4]Vec2
	center   Vec2
}

func NewQuadrangle2(v []Vec2) Quadrangle2 {
	return Quadrangle2{
		Vertices: [4]Vec2{v[0], v[1], v[2], v[3]},
		center:   v[0].Add(v[1]).Add(v[2]).Add(v[3]).Mul(0.25),
	}
}

func (q *Quadrangle2) HasPoint(p Vec2) bool {
	for k := 0; k < 4; k++ {
		l := k + 1
		if l == 4 {
			l = 0
		}
		norm := Vec2{
			q.Vertices[l].Y - q.Vertices[k].Y,
			q.Vertices[k].X - q.Vertices[l].X,
		}.Ort()
		c := norm.Dot(q.Vertices[k])
		distPt := norm.Dot(p) - c
		distCen := norm.Dot(q.center) - c
		if distCen*distPt < 0 {
			return false
		}
	}
	return true
}
