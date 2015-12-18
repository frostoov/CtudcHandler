package nevod

import (
	"encoding/binary"
	"io"
)

const (
	//HDR_W Для записи конфигурации и мониторинга ДЕКОР
	HDR_W = 0
	//ID_CONFIG Конфигурация ДЕКОР
	ID_CONFIG = 0
	//ID_MONIT Мониторинг ДЕКОР
	ID_MONIT = 1
	//ID_EVENT Не используется
	ID_EVENT = 2
	//ID_NOISE Шумы ДЕКОР
	ID_NOISE        = 3
	CFGNVD_W        = 1
	MONPDS_W        = 2
	MONPDSL_W       = 3
	MONAMPL_W       = 4
	MONSHUMTV_W     = 5
	MONBEK_W        = 6
	EVENT_W         = 7
	MONPDS_SCT_W    = 8
	MONSHUMTV_SCT_W = 9
	MONBEP_W        = 10
)

//TDateTimeKadr Структура для хранения даты и времени
type DateTime struct {
	Hsecond uint8
	Second  uint8
	Minute  uint8
	Hour    uint8
	Day     uint8
	Month   uint8
	Year    uint16
}

//SMonADC Данные мониторинга одного БЭКа
type SMonADC struct {
	ToSave  uint16           //Флаг наличия новых данных, 1 - надо их сохранить.
	Date    DateTime         //Дата и Время измерения (UTC)
	MaskPMT [4]uint16        //[ksm] маска ФЭУ измерений
	Nstarts uint16           //Кoличество запусков измерения
	Nsum    [4][6][2]uint16  //[ksm][pmt][dinod12, dinod9] кол-во измерений в спектре Пьедесталов
	Sred    [4][6][2]float32 //[ksm][pmt][dinod12, dinod9] Пьедесталы
	Sigma   [4][6][2]float32 //[ksm][pmt][dinod12, dinod9] сигма пьедесталов
}

//SMonShumTV
type SMonShumTV struct {
	ToSave   uint16        //Флаг наличия новых данных, 1 - надо их сохранить.
	Date     DateTime      //Дата и Время измерения (UTC)
	MaskPMT  [4]uint16     //[ksm] маска ФЭУ измерений
	NoisePMT [4][6]float32 //[ksm][pmt] Шумы ФЭУ в кГц
	Tbek     [4]float32    //Температура в БЭК: Tout, T3,T2,T1
	Vbek     [5]float32    //Напряжения  в БЭК: V1,V2,Vcc,V3,V4
}

//SMonBek
type SMonBek struct {
	ToSave   uint16    //Флаг наличия новых данных, 1 - надо их сохранить.
	Date     DateTime  //Время измерения шумов триггеров(UTC)
	MaskTrA  uint16    //Маска КСМ измерения шумов триггера A
	MaskTrB  uint16    //Маска КСМ измерения шумов триггера B
	MaskTrC  uint16    //Маска КСМ измерения шумов триггера C
	NoiseTrA [4]uint16 //[ksm] Шумы триггерного сигнала A
	NoiseTrB [4]uint16 //[ksm] Шумы триггерного сигнала B
	NoiseTrC [4]uint16 //[ksm] Шумы триггерного сигнала C
}

//SEvtBek
type SEvtBek struct {
	IdBek    [2]int16        //[0]-Индикатор запроса данных события данного БЭК, [1]- КСМ(0) или СКТ(1)
	MaskKSM  int16           //Маска используемых KSM
	MaskHit  [4]int16        //[ksm] маска сработавших ФЭУ
	Acp      [4][6][2]uint16 //[ksm][pmt][12d,9d] Коды АЦП
	FifoA    [4]uint16       //[ksm]
	FifoB    [4]uint16       //[ksm]
	FifoC    [4]uint16       //[ksm]
	MaskTrig [4]uint16       //[ksm] маска битов триггеров, в которых КСМ участвовал, бит:0-A,1-B,2-C
}

//SMonAdcSct Данные мониторинга одного БЭПа
type SMonAdcSct struct {
	ToSave  uint16        //Флаг наличия новых данных, 1 - надо их сохранить.
	Date    DateTime      //Дата и Время измерения (UTC)
	MaskPMT [8]uint8      //[ksm] маска ФЭУ измерений
	Nstart  uint16        //Кoличество запусков измерения
	Nsum    [8][5]uint16  //[ksm][pmt] кол-во измерений в спектре Пьедесталов
	Sred    [8][5]float32 //[ksm][pmt] Пьедесталы
	Sigma   [8][5]float32 //[ksm][pmt] сигма пьедесталов
}

//SMonShumTvSct
type SMonShumTvSct struct {
	ToSave   uint16        //Флаг наличия новых данных, 1 - надо их сохранить.
	Date     DateTime      //Дата и Время измерения (UTC)
	MaskPmt  [8]uint8      //[ksm] маска ФЭУ измерений
	NoisePmt [8][5]float32 //[ksm][pmt] Шумы ФЭУ в кГц
	Tbek     [4]float32    //Температура в БЭК: Tout, T3,T2,T1
	Vbek     [5]float32    //Напряжения  в БЭК: V1,V2,Vcc,V3,V4
}

//SMonBep
type SMonBep struct {
	ToSave     uint16       //Флаг наличия новых данных, 1 - надо их сохранить.
	Date       DateTime     //Время измерения шумов триггеров(UTC)
	MaskaTrSCT [8]uint16    //[ksm] Маска ФЭУ измерения шумов триггера СКТ
	NoiseTrSCT [8][5]uint16 //[ksm][pmt] Шумы триггерного сигнала СКТ
}

//SEvtBep
type SEvtBep struct {
	IdBek    [2]int16     //[0]-Индикатор запроса данных события данного БЭП, [1]- КСМ(0) или СКТ(1)
	MaskKsm  int16        //Маска используемых KSM
	MaskHit  [8]uint8     //[ksm] маска сработавших ФЭУ
	Acp      [8][5]uint16 //[ksm][pmt] Коды АЦП
	FifoSCT  [8][5]uint16 //[ksm][pmt]
	MaskTrig [8]uint8     //[ksm] битовая маска, участвовавших в триггере СКТ счётчиков
}

//HEADER_REC Заголовок записи NAD
type RecordHeader struct {
	Start   [5]uint8 // слово "start" - метка начала записи.
	RecType uint8    //тип записи: 0,1 - конфигурация НЕВОДа. 2,3,4,5,6 - данные мониторирования НЕВОДа. 7 - данные события.
	Date    DateTime //Дата и время (UTC)
	DataLen uint32   //Длина следующих за зоголовком данных в байтах
}

//SCONFIG_DAT Или для ДЕКОР при tip_zap==0 идентификатор ID_CONFIG...ID_NOISE
type SCONFIG_DAT struct {
	ConfBek  [32]ConfBek   //Конфигурация БЭК в НЕВОДе
	ConfTrig [8]CONFIG_TRG //Конфигурация триггерных плат
	ConfBep  [2]ConfBep    //Конфигурация БЭП в НЕВОДе
}

type SMONIT_DAT struct {
	MaskBek uint        //битовая Маска присутствующих в данных БЭК или БЭП
	Nbek    int16       //Количество присутствующих в данных БЭК или БЭП
	MonPds  [32]SMonADC //Результаты мониторинга пьедесталов БЭК
}

//NevodEvent
type NevodEvent struct {
	Meta struct {
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
		AllTime     [2]uint32    //- [0]=10000000; 100 нсек генератор тиков
		Pressure    uint32       //Давление
		Temperature uint32       //Температура
		IDDecor     uint32       //Признак наличия в данных ДЕКОРа (резерв)
		StatusReg   [8][2]uint16 //Содержимое статусных регистров без купюр
		MaskBek     uint32       // битовая Маска присутствующих в данных БЭК
		MaskBep     uint32       // битовая Маска присутствующих в данных БЭП (только два первых бита)
		Nbek        int16        //Количество присутствующих в данных БЭК (до 30 без БЭП)
		Nbep        int16        //Количество присутствующих в данных БЭП (до 2)
	}
	EventBek [32]SEvtBek //Комбинированные данные одного события от БЭК
	EventBep [2]SEvtBep  //Комбинированные данные одного события от БЭП
}

func (e *NevodEvent) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &e.Meta); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, e.EventBek[:e.Meta.Nbek]); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, e.EventBep[:e.Meta.Nbep]); err != nil {
		return err
	}
	return nil
}
