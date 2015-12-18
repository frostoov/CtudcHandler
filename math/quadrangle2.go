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

func getT(pt, p, v Vec2) float64 {
	switch {
	case v.X != 0:
		return (pt.X - p.X) / v.X
	case v.Y != 0:
		return (pt.Y - p.Y) / v.Y
	}
	panic("getT null vector")
}

func (q *Quadrangle2) Cross(l Line2) (crosses []Vec2) {
	for i := 0; i < 4; i++ {
		j := (i + 1) % 4
		ol := NewLine2Points(q.Vertices[i], q.Vertices[j])
		if crossPoint, err := l.Cross(ol); err == nil {
			pt, vec := ol.Vectors()
			cT := getT(crossPoint, pt, vec)
			iT := getT(q.Vertices[i], pt, vec)
			jT := getT(q.Vertices[j], pt, vec)
			if iT <= cT && cT <= jT || jT <= cT && cT <= iT {
				crosses = append(crosses, crossPoint)
			}
		}
	}
	return
}

func (q *Quadrangle2) HasPoint(p Vec2) bool {
	for k := 0; k < 4; k++ {
		l := (k + 1) % 4
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
