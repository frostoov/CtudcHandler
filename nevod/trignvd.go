package nevod

//TriggerConfig TODO
type TriggerConfig struct {
	Index      uint16    //Порядковый номер в конфигурации
	Base       uint32    //базовый адрес (если 0, то не используется)
	Name       [8]uint8  //имя регистра
	IDmaster   uint16    // 1- мастерный триггер, по нему ловим события
	OutMask    uint16    //Маска внешних триггеров
	ThresSum   uint16    //Пороговая сумма
	GateW      uint16    //Ширина ворот // Для СКТ ширина ворот ожидания на отдельных плоскостях
	TimerNoise uint32    //Время измерения шумов в единицах 100 нсек
	IDmask     uint16    // 0-всё закрыто, 1- по конфигурации БЭК, 2-всё открыто
	Mode       uint16    //режим работы блока
	D1         uint16    //задержка 1 в единицах 10 нсек
	D2         uint16    //задержка 2 в единицах 10 нсек
	D3         uint16    //задержка 3 в единицах нсек
	D4         uint16    //задержка 4 в единицах нсек
	TrigMaska  [8]uint16 //128 бит маска входных сигналов
	GateUD     uint16    //ширина ворот ожидания сигнала между плоскостями СКТ
	TresSumU   uint16    //Пороговая сумма для верхней плоскостей СКТ
	TresSumD   uint16    //Пороговая сумма для нижней плоскостей СКТ
}
