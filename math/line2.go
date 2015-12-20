package math

import (
	"errors"
)

//Line2 представляет линию в двухмерном простарнстве ( Ortho*r + Dist = 0).
type Line2 struct {
	Ortho Vec2
	Dist  float64
}

//NewLine2 создает линию по вектору orth и числу dist (ortho*r + dist = 0)
func NewLine2(orth Vec2, dist float64) Line2 {
	return Line2{
		Ortho: orth,
		Dist:  dist,
	}
}

//NewLine2KB создает линию по числа k и b (y=k*x+b)
func NewLine2KB(k, b float64) Line2 {
	ortho := Vec2{X: -k, Y: 1}.Ort()
	dist := ortho.Y * b
	return Line2{
		Ortho: ortho,
		Dist:  dist,
	}
}

//NewLine2Points создает Line2 по двум точкам, через которые проходит прямая.
func NewLine2Points(pt1, pt2 Vec2) Line2 {
	vec := pt1.Sub(pt2)

	orth := vec.Ortho().Ort()
	dist := -orth.Dot(pt2)
	return Line2{
		Ortho: orth,
		Dist:  dist,
	}
}

//NewLine2Vec создает Line2 по точке pt, через которую проходит прямая, и напаравляющему вектору vec прямой.
func NewLine2Vec(pt, vec Vec2) Line2 {
	orth := vec.Ortho().Ort()
	dist := -orth.Dot(pt)
	return Line2{
		Ortho: orth,
		Dist:  dist,
	}
}

//Cross возращает точку пересечения прямых l и ol.
func (l *Line2) Cross(ol Line2) (Vec2, error) {
	t := ol.Ortho.Y - ol.Ortho.X/l.Ortho.X*l.Ortho.Y
	if t == 0 {
		return Vec2{0, 0}, errors.New("Dummy")
	}
	y := -(ol.Ortho.X/l.Ortho.X*l.Dist + ol.Dist) / t
	x := (l.Dist - l.Ortho.Y*y) / l.Ortho.X
	return Vec2{X: x, Y: y}, nil
}

//Vectors возращает точку, через которую проходит l, и напаравляющий вектор l.
func (l *Line2) Vectors() (point Vec2, vector Vec2) {
	vector = l.Ortho.Ortho().Ort()
	switch {
	case l.Ortho.Y != 0:
		point = Vec2{X: 0, Y: -l.Dist / l.Ortho.Y}
	case l.Ortho.X != 0:
		point = Vec2{Y: 0, X: -l.Dist / l.Ortho.X}
	default:
		panic("Vectors null vector")
	}
	return
}

func (l *Line2) K() float64 {
	return -l.Ortho.X / l.Ortho.Y
}

func (l *Line2) B() float64 {
	return l.Dist / l.Ortho.Y
}
