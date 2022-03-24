package source

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
)

type CsvScanner struct {
	Path string

	file   *os.File
	reader *csv.Reader
	record []string
	err    error
}

func (s *CsvScanner) Open() error {
	s.file, s.err = os.Open(s.Path)
	if s.err != nil {
		return fmt.Errorf("os.Open Path %s err: %w", s.Path, s.err)
	}

	s.reader = csv.NewReader(s.file)
	return nil
}

func (s *CsvScanner) Next() bool {
	if s.err != nil {
		return false
	}

	record, err := s.reader.Read()
	if err != nil {
		if err == io.EOF {
			s.err = nil
		} else {
			s.err = err
		}
		return false
	}

	s.record = record
	return true
}

func (s *CsvScanner) Record() []string {
	return s.record
}

func (s *CsvScanner) Close() error {
	return s.file.Close()
}

func (s *CsvScanner) Err() error {
	return s.err
}

func FromCsv(path string, fun Action, options ...ScanOptions) error {
	return FromScanner(&CsvScanner{Path: path}, fun, options...)
}
