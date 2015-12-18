package math

type CoordSystem struct {
	offset Vec3
	v1     Vec3
	v2     Vec3
	v3     Vec3
}

func NewCoordSystem(offset, ox, oy, oz Vec3) CoordSystem {
	return CoordSystem{
		offset: offset,
		v1:     ox,
		v2:     oy,
		v3:     oz,
	}
}

func (c *CoordSystem) ConvertVector(v Vec3) Vec3 {
	return c.rotate(c.shift(v))
}

func (c *CoordSystem) ConvertLine(l Line3) Line3 {
	return Line3{
		c.rotate(l.Vector),
		c.ConvertVector(l.Point),
	}
}

func (c *CoordSystem) rotate(v Vec3) Vec3 {
	return Vec3{
		v.Dot(c.v1),
		v.Dot(c.v2),
		v.Dot(c.v3),
	}
}

func (c *CoordSystem) shift(v Vec3) Vec3 {
	return v.Sub(c.offset)
}
