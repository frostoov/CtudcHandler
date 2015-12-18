package math

type Line3 struct {
	Point  Vec3
	Vector Vec3
}

func NewLine3(pt1, pt2 Vec3) Line3 {
	return Line3{pt1, pt2.Sub(pt1)}
}
