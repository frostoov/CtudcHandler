package trek

import (
	"encoding/binary"
	"io"
	"time"
)

// ChamTimes содержит измерения одной камеры в формате [wire][row]time.
type ChamTimes [4][]uint

// Event содержит информацию об одном событии КТУДК.
type Event struct {
	nRun   uint64
	nEvent uint64
	time   time.Time
	hits   []Hit
}

func (e *Event) Copy() Event {
	hits := make([]Hit, len(e.hits))
	copy(hits, e.hits)
	return Event{
		nRun:   e.nRun,
		nEvent: e.nEvent,
		time:   e.time,
		hits:   hits,
	}
}

// Nrun возвращает номер рана события КТУДК.
func (e *Event) Nrun() uint {
	return uint(e.nRun)
}

// Nevent возвращает номер события КТУДК.
func (e *Event) Nevent() uint {
	return uint(e.nEvent)
}

// Time возвращает время события КТУДК.
func (e *Event) Time() time.Time {
	return e.time
}

// Hits возвращает массив хитов события КТУДК.
func (e *Event) Hits() []Hit {
	return e.hits
}

// Times возвращает измерения со всей установки в формате [chamber]*ChamTimes.
func (e *Event) Times() map[uint]*ChamTimes {
	times := make(map[uint]*ChamTimes)
	for _, h := range e.hits {
		if times[h.Chamber()] == nil {
			times[h.Chamber()] = new(ChamTimes)
		}
		times[h.Chamber()][h.Wire()] = append(times[h.Chamber()][h.Wire()], h.Time())
	}
	return times
}

// ChamberTimes возвращает измерения с камеры cham.
func (e *Event) ChamberTimes(cham uint) *ChamTimes {
	var times ChamTimes
	for _, h := range e.hits {
		if h.Chamber() == cham {
			times[h.Wire()] = append(times[h.Wire()], h.Time())
		}
	}
	return &times
}

// TriggeredChambers возвращает множество всех сработавших камер.
func (e *Event) TriggeredChambers() map[uint]bool {
	trigChams := make(map[uint]bool)
	for _, hit := range e.hits {
		trigChams[hit.Chamber()] = true
	}
	return trigChams
}

//Marshal осуществляет бинарынй маршалинг данных события в w.
func (e *Event) Marshal(w io.Writer) error {
	if err := binary.Write(w, binary.LittleEndian, e.nRun); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, e.nEvent); err != nil {
		return err
	}
	if err := e.marshalTime(w); err != nil {
		return err
	}
	if err := binary.Write(w, binary.LittleEndian, uint32(len(e.hits))); err != nil {
		return err
	}
	for i := range e.hits {
		if err := e.hits[i].Marshal(w); err != nil {
			return err
		}
	}
	return nil
}

//Unmarshal осуществляет бинарынй анмаршалинг данных события в r.
func (e *Event) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &e.nRun); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, &e.nEvent); err != nil {
		return err
	}
	if err := e.unmarhsalTime(r); err != nil {
		return err
	}
	var size uint32
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return err
	}
	if int(size) < cap(e.hits) {
		e.hits = e.hits[:size]
	} else {
		e.hits = make([]Hit, int(size))
	}
	for i := range e.hits {
		if err := e.hits[i].Unmarshal(r); err != nil {
			return err
		}
	}
	return nil
}

func (e *Event) marshalTime(w io.Writer) error {
	millis := int64(e.time.UnixNano() / 1000000)
	if err := binary.Write(w, binary.LittleEndian, millis); err != nil {
		return err
	}
	return nil
}

func (e *Event) unmarhsalTime(r io.Reader) error {
	var millis int64
	if err := binary.Read(r, binary.LittleEndian, &millis); err != nil {
		return err
	}
	sec := millis / 1000
	nsec := (millis % 1000) * 1000000
	e.time = time.Unix(sec, nsec)
	return nil
}
