package trek

import (
	"bufio"
	"errors"
	"io"
	"time"
)

// Scanner осуществляет последовательное считывание событий КТУДК.
type Scanner struct {
	header string
	reader *bufio.Reader
	event  Event
	err    error
}

func readHeader(r *bufio.Reader) (string, error) {
	header, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	header = header[:len(header)-2]
	if header != "TDS" {
		return "", errors.New("NewScanner invalid header")
	}
	return header, nil
}

// NewScanner возвращает новый Scanner, читающиц из r.
func NewScanner(r io.Reader) (*Scanner, error) {
	reader := bufio.NewReader(r)
	header, err := readHeader(reader)
	if err != nil {
		return nil, err
	}
	return &Scanner{
		header: header,
		reader: reader,
	}, nil
}

func (s *Scanner) Reset(r io.Reader) error {
	s.reader.Reset(r)
	header, err := readHeader(s.reader)
	if err != nil {
		return nil
	}
	s.header = header
	s.event = Event{0, 0, time.Now(), nil}
	s.err = nil
	return nil
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
