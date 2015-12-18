package trek

import (
	"bufio"
	"io"
)

// Scanner осуществляет последовательное считывание событий КТУДК.
type Scanner struct {
	header string
	reader *bufio.Reader
	event  Event
	err    error
}

// NewScanner возвращает новый Scanner, читающиц из r.
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
// метод Err возвращает ошибку(если ошибка - io.EOF, возвращает nil).
func (s *Scanner) Scan() bool {
	err := s.event.Unmarshal(s.reader)
	if err == nil {
		return true
	}
	if err != io.EOF {
		s.err = err
	}
	return false
}

// Record Возвращает последнее прочитанное событие.
func (s *Scanner) Record() *Event {
	return &s.event
}

// Header Возвращает заголовок данных
func (s *Scanner) Header() string {
	return s.header
}

// Err Возвращает ошибку, произошедшую при чтении.
func (s *Scanner) Err() error {
	return s.err
}
