package monitor

import "os"

type StderrShim struct {
	writer *os.File
}

func (s *StderrShim) Write(p []byte) (int, error) {
	go func(data []byte) {
		GetMessageQueue() <- string(data)
	}(p)
	return s.writer.Write(p)
}

func NewStderrShim() *StderrShim {
	return &StderrShim{
		writer: os.Stderr,
	}
}
