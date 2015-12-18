package trek

import (
	"encoding/binary"
	"github.com/frostoov/CtudcHandler/math"
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
	Nevod struct {
		Nevent      uint32       // номер события.
		Nrun        uint32       // номер рана.
		TrigNvd     uint16       // Триггеры НЕВОДа
		Nlam        int16        // кол-во сработавших модулей.
		NlamSKT     int16        // кол-во сработавших модулей СКТ.
		NfifoA      int16        // кол-во модулей с FIFO TrA.
		NfifoB      int16        // кол-во модулей с FIFO TrB.
		NfifoC      int16        // кол-во модулей с FIFO TrC.
		NfifoSKT    int16        // кол-во модулей с FIFO TrSKT.
		WaitTime    uint32       // время ожидания этого события в 100 нсек.
		AllTime     [2]uint32    // [0]=10000000; 100 нсек генератор тиков
		Pressure    uint32       // Давление
		Temperature uint32       // Температура
		IDDecor     uint32       // Признак наличия в данных ДЕКОРа (резерв)
		StatusReg   [8][2]uint16 // Содержимое статусных регистров без купюр
		MaskBek     uint32       // битовая Маска присутствующих в данных БЭК
		MaskBep     uint32       // битовая Маска присутствующих в данных БЭП (только два первых бита)
		Nbek        int16        //Количество присутствующих в данных БЭК (до 30 без БЭП)
		Nbep        int16        //Количество присутствующих в данных БЭП (до 2)
	}
	// Треки ДЕКОР
	Decor []DecorTrack
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
	e.Decor = make([]DecorTrack, l)
	if err := binary.Read(r, binary.LittleEndian, e.Decor); err != nil {
		return err
	}
	return nil
}
