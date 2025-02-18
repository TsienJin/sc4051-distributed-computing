package logging

import "os"

type StderrCustomShim struct {
	target chan string
	writer *os.File
}

func (s *StderrCustomShim) Write(p []byte) (int, error) {
	go func(data []byte) {
		s.target <- string(data)
	}(p)
	return s.writer.Write(p)
}

func NewStderrCustomShim(target chan string) *StderrCustomShim {
	return &StderrCustomShim{
		target: target,
		writer: os.Stderr,
	}
}
