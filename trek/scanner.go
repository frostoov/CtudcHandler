package trek

import (
	"bufio"
	"io"
)
// Scanner осуществляет последовательное считывание событий КТУДК
type Scanner struct {
	header string
	reader *bufio.Reader
	event  Event
	err    error
}
// NewScanner возвращает новый Scanner, читающиц из r
func NewScanner(r io.Reader) (*Scanner, error) {
	reader := bufio.NewReader(r)
	header, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	header = header[:len(header)-2]
	return &Scanner{
		header: header,
		reader: reader,
	}, nil
}
// Scan считывает следующее событие. В случае успеха возвращает true,
// при возникновении ошибки возвращает false. После того как Scan возвращает false,
// метод Err возвращает ошибку(если ошибка - io.EOF, возвращает nil)
func (e *Scanner) Scan() bool {
	if err := e.event.Unmarshal(e.reader); err == nil {
		return true
	} else {
		if err != io.EOF {
			e.err = err
		}
		return false
	}
}
// Record Возвращает последнее прочитанное событие
func (s *Scanner) Record() *Event {
	return &s.event
}
// Record Возвращает заголовок данных
func (s *Scanner) Header() string {
	return s.header
}

func (s *Scanner) Err() error {
	return s.err
}
