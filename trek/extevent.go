package trek

import (
	"encoding/binary"
	"io"
	"time"

	"github.com/frostoov/CtudcHandler/math"
	"github.com/frostoov/CtudcHandler/nevod"
)

// DecorTrack содержит данные о треке ДЕКОР.
type DecorTrack struct {
	//Тип трека: 0 - long, 1 - ShSh
	Type int8
	//Прямая трека
	Track math.Line3
}

// ExtHeader содержит метаданные рана
type ExtHeader struct {
	// Номер первого события рана
	FirstEvent uint64
	// Номер последнего события рана
	LastEvent uint64
	// Время начала рана
	StartTime time.Time
	// Время стопа рана
	StopTime time.Time
	// "Живая" длительность рана
	LiveDur time.Duration
	// Полная длительность рана
	FullDur time.Duration
}

// Marshal осуществляет бинарный маршалинг заголовка в w.
func (e *ExtHeader) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, e.FirstEvent); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.LastEvent); err != nil {
		return err
	}
	if err := marshalTime(w, e.StartTime); err != nil {
		return err
	}
	if err := marshalTime(w, e.StopTime); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, int64(e.LiveDur.Seconds())); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, int64(e.FullDur.Seconds())); err != nil {
		return err
	}
	return nil
}

// Unmarshal осуществляет бинарный анмаршалинг заголовка из r.
func (e *ExtHeader) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &e.FirstEvent); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.LastEvent); err != nil {
		return err
	}
	startTime, err := unmarhsalTime(r)
	if err != nil {
		return err
	}
	e.StartTime = startTime
	endTime, err := unmarhsalTime(r)
	if err != nil {
		return err
	}
	e.StopTime = endTime
	var liveDur int64
	if err := binary.Read(r, binary.LittleEndian, &liveDur); err != nil {
		return err
	}
	e.LiveDur = time.Duration(liveDur) * time.Second
	var fullDur int64
	if err := binary.Read(r, binary.LittleEndian, &fullDur); err != nil {
		return err
	}
	e.FullDur = time.Duration(fullDur) * time.Second
	return nil

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
