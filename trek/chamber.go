package trek

import (
	"math"

	geo "github.com/frostoov/CtudcHandler/math"
)

// ChamberDesc содержит описание дрейфовой камеры.
type ChamberDesc struct {
	//Точки дрейфовой камеры.
	Points [3]geo.Vec3 `json:"points"`
	//Оффсет времен дрейфа для каждой проволки.
	Offsets [4]uint `json:"offsets"`
	//Скорость дрейфа для каждой проволки.
	Speeds [4]float64 `json:"speeds"`
	//Номер плоскости дрейфовой камеры.
	Plane uint `json:"plane"`
	//Номер группы дрейфовой камеры.
	Group uint `json:"group"`
	//Номер камеры.
	Number uint `json:"number"`
}

// TrackDesc содержит описание реконструированного трека.
type TrackDesc struct {
	// Прямая трека.
	Line geo.Line2
	// Точки, по которым был восстановлен трека.
	Points [4]geo.Vec2
	// Отклонение прямой.
	Deviation float64
	// Вермена с TDC.
	Times [4]uint
}

const (
	chamberWidth  = 500
	chamberHeight = 112
	chamberLength = 4000
)

// Chamber представляет дрейфовую камеру.
type Chamber struct {
	desc  ChamberDesc
	coord geo.CoordSystem
	hex   geo.Hexahedron
}

// NewChamber создает Chamber по описанию chamDesc.
func NewChamber(chamDesc ChamberDesc) *Chamber {
	return &Chamber{
		desc:  chamDesc,
		coord: mkChamberCoord(chamDesc.Points[:]),
		hex:   mkChamberHexahedron(chamDesc.Points[:]),
	}
}

// LineProjection проецирует прямую на фронтальную плоскость камеры.
func (c *Chamber) LineProjection(l geo.Line3) geo.Line2 {
	l = c.coord.ConvertLine(l)
	return geo.NewLine2Vec(geo.Vec2{X: l.Point.X, Y: l.Point.Y}, geo.Vec2{X: l.Vector.X, Y: l.Vector.Y})
}

// CreateTrack реконструирует трек по измерениям с камеры.
func (c *Chamber) CreateTrack(times *ChamTimes) *TrackDesc {
	return mkTrackDesc(times, &c.desc)
}

// Hexahendron возвращает геометрическое представление камеры.
func (c *Chamber) Hexahendron() *geo.Hexahedron {
	return &c.hex
}

// Number возвращает номер камеры.
func (c *Chamber) Number() uint {
	return c.desc.Number
}

// Plane возвращает номер плоскости камеры.
func (c *Chamber) Plane() uint {
	return c.desc.Plane
}

// Group возвращает номер группы камеры.
func (c *Chamber) Group() uint {
	return c.desc.Group
}

// Offsets возвращает оффсеты для каждой проволки камеры.
func (c *Chamber) Offsets() []uint {
	return c.desc.Offsets[:]
}

// Speeds возвращает скорости дрейфа для каждой проволки камеры.
func (c *Chamber) Speeds() []float64 {
	return c.desc.Speeds[:]
}

// Width возвращает ширину дрейфовой камеры.
func (c *Chamber) Width() float64 {
	return chamberWidth
}

// Height возвращает высоту дрейфовой камеры.
func (c *Chamber) Height() float64 {
	return chamberHeight
}

// Length возвращает длину дрейфовой камеры.
func (c *Chamber) Length() float64 {
	return chamberLength
}

func mkChamberCoord(pts []geo.Vec3) geo.CoordSystem {
	ox := pts[1].Sub(pts[0]).Ort()
	oz := pts[2].Sub(pts[0]).Ort()
	oy := ox.Cross(oz)
	return geo.NewCoordSystem(pts[0], ox, oy, oz)
}

func mkChamberHexahedron(pts []geo.Vec3) geo.Hexahedron {
	const d = chamberWidth / 2

	//Вспомогательные векторыX:
	p13 := pts[2].Sub(pts[0])
	p12 := pts[1].Sub(pts[0])
	w := p12.Cross(p13).Ort().Mul(d)
	vertices := []geo.Vec3{
		pts[0].Add(w),
		pts[0].Sub(w),
		pts[1].Sub(w),
		pts[1].Add(w),
		pts[2].Add(w),
		pts[2].Sub(w),
		pts[2].Sub(w).Add(p12),
		pts[2].Add(w).Add(p12),
	}
	return geo.NewHexahedron(vertices)
}

var wires = [4]geo.Vec2{
	geo.Vec2{X: 41, Y: 0.75},
	geo.Vec2{X: 51, Y: -0.75},
	geo.Vec2{X: 61, Y: 0.75},
	geo.Vec2{X: 71, Y: -0.75},
}

// ChamDists содержит длины дрейфа в одной камеры в формате [wire][row]dist.
type ChamDists [4][]float64

// TrackDists содержит длины дрейфа частиц по которым был востановлен трек.
type TrackDists [4]float64

// TrackTimes содержит измерения по которым был востановлен трек.
type TrackTimes [4]uint

func mkTrackDesc(times *ChamTimes, chamDesc *ChamberDesc) *TrackDesc {
	dists := mkChamDists(times, chamDesc)
	depth := getDepth(dists)
	if depth != uint(1) {
		return nil
	}
	desc := TrackDesc{
		Deviation: math.Inf(1),
	}

	var ind [4]int
	var tmpDesc TrackDesc
	for ind[0] = range dists[0] {
		for ind[1] = range dists[1] {
			for ind[2] = range dists[2] {
				for ind[3] = range dists[3] {
					trackDists := mkTrackDists(dists, &ind)
					if mkTrack(&trackDists, &tmpDesc) && tmpDesc.Deviation < desc.Deviation {
						tmpDesc.Times = mkTrackTimes(times, &ind)
						desc = tmpDesc
					}
				}
			}
		}
	}
	if desc.Deviation != math.Inf(1) && systemError(&desc) {
		return &desc
	}
	return nil
}

func mkTrack(dists *TrackDists, desc *TrackDesc) bool {
	points := [4]geo.Vec2{wires[0], wires[1], wires[2], wires[3]}
	var line geo.Line2
	desc.Deviation = math.Inf(1)
	numPermutations := uint(math.Pow(2, float64(len(dists))))

	for i := uint(0); i < numPermutations; i++ {
		//Изменяем знаки на противоположные
		for j := uint(0); j < uint(len(dists)); j++ {
			if i&(1<<j) != 0 {
				points[j].Y = -dists[j]
			} else {
				points[j].Y = dists[j]
			}
			points[j].Y += wires[j].Y
		}
		dev := leastSquares(points[:], &line)
		if dev != -1 && dev < desc.Deviation {
			desc.Deviation = dev
			desc.Line = line
			desc.Points = points
		}
	}
	return desc.Deviation != math.Inf(1)
}

func mkTrackTimes(chamTimes *ChamTimes, ind *[4]int) TrackTimes {
	var times TrackTimes
	for i := range times {
		times[i] = chamTimes[i][ind[i]%len(chamTimes[i])]
	}
	return times
}

func mkTrackDists(chamDists *ChamDists, ind *[4]int) TrackDists {
	var dists TrackDists
	for i := range dists {
		dists[i] = chamDists[i][ind[i]%len(chamDists[i])]
	}
	return dists
}

func mkChamDists(times *ChamTimes, desc *ChamberDesc) *ChamDists {
	var dists ChamDists
	for wire := range times {
		for _, time := range times[wire] {
			offset := desc.Offsets[wire]
			if time > offset {
				speed := desc.Speeds[wire]
				dists[wire] = append(dists[wire], float64(time-offset)*speed)
			}
		}
	}
	return &dists
}

func getDepth(dists *ChamDists) uint {
	depth := uint(math.MaxUint64)
	for _, wireDists := range dists {
		if uint(len(wireDists)) < depth {
			depth = uint(len(wireDists))
		}
	}
	return depth
}

func leastSquares(pts []geo.Vec2, line *geo.Line2) float64 {
	if len(pts) < 2 {
		return -1
	}
	var sumX, sumY, sumXY, sumXX float64
	for i := range pts {
		sumX += pts[i].X
		sumY += pts[i].Y
		sumXY += pts[i].X * pts[i].Y
		sumXX += pts[i].X * pts[i].X
	}
	l := float64(len(pts))
	exp := l*sumXX - sumX*sumX
	if exp != 0 && math.Abs(exp) > 1e-60 {
		k := (l*sumXY - sumX*sumY) / exp
		b := (sumY - k*sumX) / l
		dev := 0.0
		for i := range pts {
			dev += math.Pow((k*pts[i].X+b)-pts[i].Y, 2)
		}
		*line = geo.NewLine2KB(k, b)
		return dev
	}
	return -1
}

func sign(val float64) float64 {
	switch {
	case val > 0:
		return 1
	case val < 0:
		return -1
	default:
		return 0
	}
}

func systemError(desc *TrackDesc) bool {
	var r float64
	for i := range desc.Points {
		trackSign := sign(desc.Points[i].Y)
		switch trackSign * sign(wires[i].Y) {
		case 1:
			if math.Abs(desc.Points[i].Y) > 6.2 {
				r = 6.2
			} else {
				r = desc.Points[i].Y
			}
		case -1:
			if math.Abs(desc.Points[i].Y) > 3.6 {
				r = 3.6
			} else {
				r = desc.Points[i].Y
			}
		default:
			return false
		}
		desc.Points[i].Y += trackSign * getSystemError(r, math.Atan(desc.Line.K()))
	}
	desc.Deviation = leastSquares(desc.Points[:], &desc.Line)
	return true
}

func getSystemError(r, ang float64) float64 {
	return r * (1/math.Cos(ang) - 1)
}
