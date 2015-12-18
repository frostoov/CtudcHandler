package math

type Hexahedron struct {
	Vertices [8]Vec3
	Polygons [6]Quadrangle3
}

func NewHexahedron(v []Vec3) Hexahedron {
	return Hexahedron{
		[8]Vec3{v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7]},
		getPolyons(v),
	}
}

func (h *Hexahedron) Cross(l Line3) []Vec3 {
	crosses := make([]Vec3, 0)
	for i := range h.Polygons {
		if c, err := h.Polygons[i].Cross(l); err == nil {
			crosses = append(crosses, c)
		}
	}
	return crosses
}

func (h *Hexahedron) Crossing(l Line3) bool {
	for i := range h.Polygons {
		if _, err := h.Polygons[i].Cross(l); err == nil {
			return true
		}
	}
	return false
}

func getPolyons(v []Vec3) [6]Quadrangle3 {
	return [6]Quadrangle3{
		NewQuadrangle3([]Vec3{v[0], v[1], v[2], v[3]}),
		NewQuadrangle3([]Vec3{v[4], v[5], v[6], v[7]}),
		NewQuadrangle3([]Vec3{v[0], v[1], v[5], v[4]}),
		NewQuadrangle3([]Vec3{v[3], v[2], v[6], v[7]}),
		NewQuadrangle3([]Vec3{v[0], v[4], v[7], v[3]}),
		NewQuadrangle3([]Vec3{v[1], v[5], v[6], v[2]}),
	}
}
