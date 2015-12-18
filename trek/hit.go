package trek

import (
	"encoding/binary"
	"fmt"
	"io"
)
// Hit содержит информацию о хите TDC
type Hit struct {
	channel uint32
	time    uint32
}
//String созадет строку с описанием хита в формате hit[номер_камер, номер_проволки]: измерение
func (h Hit) String() string {
	return fmt.Sprintf("hit[%d, %d]: %d", h.Chamber(), h.Wire(), h.Time())
}
//Time возвращает временя измерения TDC в пикосекундах
func (h Hit) Time() uint {
	return uint(h.time)
}
//Time возвращает номер проволки измерения TDC (нумерация с 0)
func (h Hit) Wire() uint {
	return uint(h.channel & 0xFF)
}
//Time возвращает номер камеры измерения TDC  (нумерация с 0)
func (h Hit) Chamber() uint {
	return uint(h.channel >> 8)
}
//Unmarshal осуществляет десериализацию данных хита в r
func (h *Hit) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &h.channel); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &h.time); err != nil {
		return err
	}
	return nil
}
//Unmarshal осуществляет сериализацию данных хита в w
func (h Hit) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, h.channel); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, h.time); err != nil {
		return err
	}
	return nil
}
