package nevod

import (
	"encoding/binary"
	"io"
)

type Scanner struct {
	reader io.ReadSeeker

	nevodData Event
	header    RecordHeader
	decorData StrDecor

	err error
}

func NewScanner(r io.ReadSeeker) *Scanner {
	return &Scanner{
		reader: r,
	}
}

func (s *Scanner) Record() *Event {
	return &s.nevodData
}

func (s *Scanner) Error() error {
	return s.err
}

func (s *Scanner) Scan() (success bool) {
	for {
		if err := binary.Read(s.reader, binary.LittleEndian, &s.header); err != nil {
			s.setError(err)
			return
		}

		switch s.header.RecType {
		case hdr:
			{
				recordType := s.header.DataLen
				switch recordType {
				case idConfig:
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.ConfCntr); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.Conf); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.TrigConf); err != nil {
						s.setError(err)
						return
					}
				case idMonit:
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.ConfMonit); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.CmonitAll); err != nil {
						s.setError(err)
						return
					}
				case idNoise:
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.Counter); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.ConfNoise); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.IDcnoise); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.LenCnoise); err != nil {
						s.setError(err)
						return
					}
					if err := binary.Read(s.reader, binary.LittleEndian, &s.decorData.Cnoise); err != nil {
						s.setError(err)
						return
					}
				}
			}
		case recordEvent:
			if err := s.nevodData.Unmarshal(s.reader); err != nil {
				s.setError(err)
				return
			}
			var lenadd [2]uint8
			s.reader.Read(lenadd[:])
			if lenadd[0] != 0 {
				if _, err := s.reader.Seek(int64(4*lenadd[0]), 1); err != nil {
					s.setError(err)
					return
				}
			}
			if lenadd[1] != 0 {
				binary.Read(s.reader, binary.LittleEndian, &s.decorData.ConfEvent)
				binary.Read(s.reader, binary.LittleEndian, &s.decorData.LenCeventAll)
				if s.decorData.LenCeventAll != 0 {
					if err := s.decorData.CeventAll.Unmarshal(s.reader); err != nil {
						s.setError(err)
						return
					}
				}
			}
			success = true
		default:
			if _, err := s.reader.Seek(int64(s.header.DataLen), 1); err != nil {
				s.setError(err)
				return
			}
		}
		var bstop [4]uint8
		if _, err := s.reader.Read(bstop[:]); err != nil || string(bstop[:]) != "stop" {
			s.setError(err)
			return
		}
		if success {
			return
		}
	}
}

func (s *Scanner) setError(err error) {
	if err != io.EOF {
		s.err = err
	}
}
