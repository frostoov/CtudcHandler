package trek

import (
	"encoding/binary"
	"fmt"
	"io"
)

// HitType Тип хита
type HitType uint8

func (h HitType) String() string {
	if h == Leading {
		return "leading"
	}
	return "trailing"
}

const (
	// Leading leading edge detection
	Leading HitType = iota
	// Trailing trailing edge detection
	Trailing
)

// Hit содержит информацию о хите TDC.
type Hit struct {
	channel uint32
	time    uint32
}

// String созадет строку с описанием хита в формате hit[номер_камер, номер_проволки]: измерение.
func (h Hit) String() string {
	return fmt.Sprintf("hit[%d, %d, %v]: %d", h.Chamber(), h.Wire(), h.Type(), h.Time())
}

//Time возвращает временя измерения TDC в пикосекундах.
func (h Hit) Time() uint {
	return uint(h.time)
}

// Wire возвращает номер проволки измерения TDC (нумерация с 0).
func (h Hit) Wire() int {
	return int(h.channel & 0xFF)
}

// Chamber возвращает номер камеры измерения TDC  (нумерация с 0).
func (h Hit) Chamber() int {
	return int((h.channel >> 8) & 0xFFFF)
}

// Type возвращает тип хита
func (h Hit) Type() HitType {
	return HitType(h.channel >> 28)
}

// Unmarshal осуществляет десериализацию данных хита в r.
func (h *Hit) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.channel); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.time); err != nil {
		return err
	}
	return nil
}

// Marshal осуществляет сериализацию данных хита в w.
func (h Hit) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, h.channel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.time); err != nil {
		return err
	}
	return nil
}
