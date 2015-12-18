package nevod

import (
	"io"
	"encoding/binary"
)

const (
	//MAXDCR максимальное количество ДЕКОР
	MAXDCR = 1
	//MAXPM максимальное количество ПМ
	MAXPM = 8
	//NCNTR кол-во контроллеров в одной ЭВМ
	NCNTR = 4
	//NCHAN кол-во каналов в одном контроллере
	NCHAN = 4
	//NGROUP8 максимально возможное кол-во групп по 8 бит
	NGROUP8 = 76
	//NGROUP16 в полной конфигурации 38 групп по 16 бит
	NGROUP16 = 38
	//NGROUP64 максимально возможное кол-во групп по 64 бита
	NGROUP64 = 10
)

//ConfPm
type ConfPm struct {
	WrkCntr int8         // NCNTR-бита: 1-найден контроллер, 0-нет
	NumCntr [NCNTR]uint8 // номера контроллеров из плат
}

//ConfChannel Структура конфигурации одного канала
type ConfChannel struct {
	Include       int16   // 0- не в работе , 1- в работе
	Nx            int16   // кол-во бит в камерах X
	Ny            int16   // кол-во бит в камерах Y
	Xx, Yx, Zx    float32 // координаты первого бита данных в камерах X
	Xy, Yy, Zy    float32 // координаты первого бита данных в камерах Y
	VXx, VYx, VZx float32 // Направляющий вектор в камерах X
	VXy, VYy, VZy float32 // Направляющий вектор в камерах Y
}

//TrigComment Структура данных из триггерной платы
type TrigComment struct {
	Type    int16     //Номер физики
	Comment [66]uint8 //комментарий
}

//TrigConf
type TrigConf struct {
	Port         uint16         // Адрес триггерной платы
	Interval     uint32         // Интервал времени для счёта шумов
	Freq         int16          // Тактовая частота генератора в МГц
	StepInterval int16          // Единица для интервала измерений шумов в нсек
	Gate         int16          // Время совпадения в нсек
	Width1       int16          // длительность1 в нсек
	Delay1       int16          // задержка1 в нсек
	Width2       int16          // длительность2 в нсек
	Delay2       int16          // задержка2 в нсек
	Width3       int16          // длительность3 в нсек
	Delay3       int16          // задержка3 в нсек
	Reserv       int16          // резерв
	TrigComment  [8]TrigComment //идентификаторы битов триггера
	Tabl         [256]uint8     //Таблица загрузки триггерных условий
}

//CMonitorAll структура данных полученная при мониторировании от всех PM
type CMonitorAll struct {
	IdCmonit int8 // битовый индикатор наличия данных от ПМ
	Conf     [MAXPM]ConfPm
	Nbad     [MAXPM][NCNTR][NCHAN]int16 // кол-во стрипов с ошибками
}

//CNoise структура данных по шумам от одной PM
type CNoise struct {
	MaskCntr int8                 // NCNTR-бита: 0-нет данных контроллера, 1-есть
	NoiseBuf [NCNTR * NCHAN]int16 // запакованные данные по NCHAN слова на каждый не нулевой бит в maskacntr
}

//DecorEvent Структура данных всего события со всех PM, это будет записываться на диск.
type DecorEvent struct {
	Meta struct {
		Start           [6]uint8     // Ключевое слово начала записи
		Type            int16        // Тип записи:0-Config,1-монитор,2-Experement event,3-Noise
		Nrun            uint32       // Номер текущего рана
		Nevent          uint32       // Номер текущего события
		Hund            uint8        // hundredths of seconds
		Sec             uint8        // seconds
		Min             uint8        // minutes
		Hour            uint8        // hours
		Day             int8         // day of the month
		Mon             int8         // month (1 = Jan)
		Year            int16        // current year
		Mcntr           uint32       //32 бита : 0-не опрашивался , 1-опрашивался
		Len             int16        // Длина области данных в байтах
		Trig            uint16       // информация по триггеру
		WeitTime        uint32       // Время ожидания этого события в единицах step
		History         [2][16]uint8 // предистория по 1/freq(в обратную сторону)
		Counter         [2][8]uint16 // живое время по 8-ок пл-тей в единицах 8*step
		MaskaLamChan    [16]uint8    // по NCHAN-бита маска LAMов на все PM
		MaskLenMaskMask uint32       // 32-бита : 0 maska_masok- char, 1 maska_masok- int16_t
		MaskCntr        uint32       //32 бита : 0-нет данных , 1-есть
		LenAllData      int16        // Длина запакованных сработавших стрипов
	}
	Buf [1514 * 8]uint8 // запакованные данные
}

func (e *DecorEvent) Unmarshal(r io.Reader) error {
	if err := binary.Read(r, binary.LittleEndian, &e.Meta); err != nil {
		return err
	}
	if err := binary.Read(r, binary.LittleEndian, e.Buf[:e.Meta.LenAllData]); err != nil {
		return err
	}
	return nil
}

//StrDecor Структура данных работы с DECOR_MAIN
type StrDecor struct {
	ErrOper        int32        // маска ошибок появившихся при выполнении операции. Если ErrOper&64 != 0, то ошибка триггерной платы и дальнейшей информации нет
	MaskPackageErr int16        //от кого есть "ERROR"
	MaskPackageBad int16        // маска запорченных пакетов
	PmConfig       int16        // требуемая конфигурация сети
	PmWork         int16        // Битовая маска рабочих машин
	PmExp          int16        // битовая маска участвующих в эксперименте PM
	IdCinit        [MAXPM]int16 // индикатор наличия данных
	LenCinit       [MAXPM]int16 // длина данных
	Cinit          [MAXPM]ConfPm
	Nrun           int32            // Номер запуска программы
	ConfCntr       uint32           // маска контроллеров по конфигурации
	Conf           [128]ConfChannel // Структура конфигурации всего ДЕКОРа
	TrigConf       [2]TrigConf      //Конфигурация для триггерной платы
	ConfMonit      uint32           // маска контроллеров по мониторингу
	CmonitAll      CMonitorAll
	Counter        [2][8]uint16 //шумы триггеров за 1 сек
	ConfNoise      uint32       // маска контроллеров по шумам
	IdCnoise       [MAXPM]int16 // индикатор наличия данных
	LenCnoise      [MAXPM]int16 // длина данных
	Cnoise         [MAXPM]CNoise
	ConfEvent      uint32    // маска опрашиваемых контроллеров по событию
	LenCeventAll   int16     // длина данных
	CeventAll      DecorEvent // Структура одного события ДЕКОРа
}
