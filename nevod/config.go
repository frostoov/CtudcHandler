package nevod

const (
	//MAXBEP Максимальное количество БЭПов.
	MAXBEP = 2
	//MAXBEK Максимальное количество БЭКов.
	MAXBEK = 32
)

//LedKsm Коды для поджигателя светодиодов КСМ.
type LedKsm struct {
	Base  uint8    //Базовое напряжение
	Range uint8    //Диапазон
	Pmt   [6]uint8 //Коды подстветок
}

//ConfKsm Структура конфигурации одного модуля (для чтения конфигурационного файла).
type ConfKsm struct {
	Enable    int8   //0- неиспользуется, 1- используется
	MaskPMT   int8   //Битовая маска используемых ФЭУ
	TrigDelay uint16 //Задержка сигнала "Хранение" в нсек
	Threshold uint16 //Порог дискриминаторов 0..255
	TrigWord  uint8  //Слово в маске триггеров (всего 16 слов - 128 бит)
	TrBit     uint8  //бит в слове в маске триггеров (всего 16 бит)
	TrigA     uint8  //участие в триггере A
	TrigB     uint8  //участие в триггере B
	TrigC     uint8  //участие в триггере C
	X, Y, Z   uint16 // Координаты положения в системе НЕВОД, в мм
	Led       LedKsm //Коды внутримодульных подсветок
}

//ConfBek Структура конфигурации БЕК из файла bek.cfg.
type ConfBek struct {
	Enable    int8       //0- неиспользуется, 1- используется
	IDbek     int8       //идентификатор: 0-КСМ или 1-СКТ
	NumberBEK int8       //Номер БЭК
	MaskKSM   int8       //Битовая маска используемых КСМ по данным CfgKSM
	ConfKSM   [4]ConfKsm //Конфигурация и параметры КСМ, подключённых к БЭК
}

//ConfKsmSct Структура конфигурации одного модуля БЭП (для чтения конфигурационного файла).
type ConfKsmSct struct {
	Enable    int8      //0- неиспользуется, 1- используется
	MaskPMT   int8      //Битовая маска используемых ФЭУ
	TrigDelay uint16    //Задержка сигнала "Хранение" в нсек
	Threshold uint16    //Порог дискриминаторов 0..255
	Ipl       [5]uint8  //Номер плоскости счётчика (4-12)
	Istr      [5]uint8  //Номер струны счётчика (1-5)
	Idownup   [5]uint8  //0-нижний, 1-верхний
	TrigWord  [5]uint8  //[pmt] Слово в маске триггеров (всего 16 слов - 128 бит)
	TrigBit   [5]uint8  // [pmt] бит в слове в маске триггеров (всего 16 бит)
	TrigSCT   [5]uint8  //[pmt] участие в триггере SCT
	X, Y, Z   [5]uint16 //[pmt] Координаты положения в системе НЕВОД, в мм
}

//ConfBep Структура конфигурации БЕП из файла bep.cfg.
type ConfBep struct {
	Enable    int8          //0- неиспользуется, 1- используется
	IDbek     int8          //идентификатор: 0-КСМ или 1-СКТ
	NumberBEK int8          //Номер БЭП
	MaskKSM   uint8         //Битовая маска используемых контроллеров по данным CfgKSM
	ConfKSM   [8]ConfKsmSct //Конфигурация и параметры контроллеров, подключённых к БЭП
}

//ConfSct Временная рабочая структура конфигурации счётчика СКТ из файла sct.cfg.
type ConfSct struct {
	Enable    int8   //0- неиспользуется, 1- используется
	Ipl       int8   //Номер плоскости (4..12)
	Istr      int8   //номер струны (1..5)
	Idownup   int8   // 0- нижняя плоскость, 1-верхняя плоскость
	X, Y, Z   uint16 // Координаты положения (центр счётчика) в системе НЕВОД, в мм
	NumberBEK int8   //Номер БЭП  (1..32)
	Icntr     int8   //номер контроллера внутри БЭП (0..7)
	Ipmt      int8   //Номер канала АЦП (0..4)
	TrigWord  uint8  //Слово в блоке триггеров SCT (всего 16 слов - 128 бит)
	TrigBit   uint8  //бит в слове в блоке триггеров SCT (всего 16 бит)
}

//TrigMask Маски для входов триггерных блоков.
type TrigMask struct {
	MaskA   [8]uint16 //Триггер А
	MaskB   [8]uint16 //Триггер B
	MaskC   [8]uint16 //Триггер C
	MaskSKT [8]uint16 //Триггер СКТ
}
