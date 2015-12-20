package trek

import (
	"encoding/binary"
	"github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/nevod"
	"io"
)

// DecorTrack содержит данные о треке ДЕКОР.
type DecorTrack struct {
	//Тип трека: 0 - long, 1 - ShSh
	Type int8
	//Прямая трека
	Track math.Line3
}

// ExtEvent содержит данные о событии с КТУДК, ДЕКОР и НЕВОД.
type ExtEvent struct {
	// Данные КТУДК
	Ctudc Event
	// Данные НЕВОД
	Nevod nevod.EventMeta
	// Треки ДЕКОР
	Decor []DecorTrack
}

// Copy создает и возвращает копию e
func (e *ExtEvent) Copy() ExtEvent {
	decor := make([]DecorTrack, len(e.Decor))
	copy(decor, e.Decor)
	return ExtEvent{
		Ctudc: e.Ctudc.Copy(),
		Nevod: e.Nevod,
		Decor: decor,
	}
}

// Marshal осуществляет бинарный маршалинг события в w.
func (e *ExtEvent) Marshal(w io.Writer) error {
	if err := e.Ctudc.Marshal(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.Nevod); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint64(len(e.Decor))); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.Decor); err != nil {
		return err
	}
	return nil
}

// Unmarshal осуществляет бинарный анмаршалинг события из r.
func (e *ExtEvent) Unmarshal(r io.Reader) error {
	if err := e.Ctudc.Unmarshal(r); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.Nevod); err != nil {
		return err
	}
	var l uint64
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return err
	}
	if int(l) < cap(e.Decor) {
		e.Decor = e.Decor[:l]
	} else {
		e.Decor = make([]DecorTrack, l)
	}
	if err := binary.Read(r, binary.LittleEndian, e.Decor); err != nil {
		return err
	}
	return nil
}
