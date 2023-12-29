package gins

import "bytes"

type Scanner struct {
	s     string
	len   int
	index int
}

func (s *Scanner) Next() bool {
	s.index++
	return s.index < s.len
}

func (s *Scanner) Read() byte {
	return s.s[s.index]
}

func (s *Scanner) String() string {
	s.len = len(s.s)
	s.index = -1
	buf := &bytes.Buffer{}
	for s.Next() {
		b := s.Read()
		if 'A' <= b && b <= 'Z' {
			buf.Write([]byte{'/', b + 32})
		} else if b == '_' && s.Next() {
			b = s.Read()
			switch b {
			case '8':
				buf.WriteString("/*")
			case '1':
				buf.WriteString("/:")
			default:
				buf.WriteByte(b)
			}
		} else {
			buf.WriteByte(b)
		}
	}
	return buf.String()
}

func NewScanner(s string) *Scanner {
	return &Scanner{
		s:     s,
		len:   len(s),
		index: -1,
	}
}

func ParseName(s string) string {
	return NewScanner(s).String()
}
