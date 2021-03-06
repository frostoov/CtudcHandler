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

func (e *Event) ChamberDepths() map[int]*[4]int {
	ds := make(map[int]*[4]int)
	for _, h := range e.hits {
		if ds[h.Chamber()] == nil {
			ds[h.Chamber()] = new([4]int)
		}
		ds[h.Chamber()][h.Wire()]++
	}
	return ds

}

func (e *Event) WireDepths(cham int) [4]int {
	ds := [4]int{}
	for _, h := range e.hits {
		if cham == int(h.Chamber()) {
			ds[h.Wire()]++
		}
	}
	return ds
}

// Nrun возвращает номер рана события КТУДК.
func (e *Event) Nrun() uint {
	return uint(e.nRun)
}

func (e *Event) SetNevent(n uint) {
	e.nEvent = uint64(n)
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
func (e *Event) Times() map[int]*ChamTimes {
	times := make(map[int]*ChamTimes)
	for _, h := range e.hits {
		if times[h.Chamber()] == nil {
			times[h.Chamber()] = new(ChamTimes)
		}
		times[h.Chamber()][h.Wire()] = append(times[h.Chamber()][h.Wire()], h.Time())
	}
	return times
}

// ChamberTimes возвращает измерения с камеры cham.
func (e *Event) ChamberTimes(cham int) *ChamTimes {
	var times ChamTimes
	for _, h := range e.hits {
		if h.Chamber() == cham {
			times[h.Wire()] = append(times[h.Wire()], h.Time())
		}
	}
	return &times
}

// TriggeredChambers возвращает множество всех сработавших камер.
func (e *Event) TriggeredChambers() map[int]bool {
	trigChams := make(map[int]bool)
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
	if err := marshalTime(w, e.time); err != nil {
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
	t, err := unmarhsalTime(r)
	if err != nil {
		return err
	}
	e.time = t
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

func marshalTime(w io.Writer, t time.Time) error {
	millis := int64(t.UnixNano() / 1000000)
	if err := binary.Write(w, binary.LittleEndian, millis); err != nil {
		return err
	}
	return nil
}

func unmarhsalTime(r io.Reader) (time.Time, error) {
	var millis int64
	if err := binary.Read(r, binary.LittleEndian, &millis); err != nil {
		return time.Time{}, err
	}
	sec := millis / 1000
	nsec := (millis % 1000) * 1000000
	return time.Unix(sec, nsec), nil
}
