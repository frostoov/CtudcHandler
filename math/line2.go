package math

import (
	"errors"
	"math"
)

//Line2 представляет линию в двухмерном простарнстве (Y = K*X + B).
type Line2 struct {
	K, B float64
}

//NewLine2Points создает Line2 по двум точкам, через которые проходит прямая.
func NewLine2Points(pt1, pt2 Vec2) Line2 {
	vec := pt1.Sub(pt2)
	k := math.Inf(1)
	if vec.X != 0 {
		k = vec.Y / vec.X
	}
	b := pt1.Y - k*pt1.X
	return Line2{k, b}
}

//NewLine2Vec создает Line2 по точке pt, через которую проходит прямая, и напаравляющему вектору vec прямой.
func NewLine2Vec(pt, vec Vec2) Line2 {
	k := vec.Y / vec.X
	b := pt.Y - k*pt.X
	return Line2{k, b}
}

//Cross возращает точку пересечения прямых l и ol.
func (l *Line2) Cross(ol Line2) (Vec2, error) {
	if l.K == ol.K {
		return Vec2{0, 0}, errors.New("Dummy")
	}
	if l.B != ol.B {
		x := (l.K - ol.K) / (ol.B - l.B)
		y := l.K*x + l.B
		return Vec2{x, y}, nil
	}
	return Vec2{0, 0}, nil
}

//Vectors возращает точку, через которую проходит l, и напаравляющий вектор l.
func (l *Line2) Vectors() (point Vec2, vector Vec2) {
	vector = Vec2{1, l.K}.Ort()
	point = Vec2{0, l.B}
	return
}
