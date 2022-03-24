package source

import (
	"bufio"
	"fmt"
	"os"
)

type TxtScanner struct {
	Path string

	file      *os.File
	s         *bufio.Scanner
	record    []string
	err       error
	splitFunc func(string) []string
}

func (t *TxtScanner) Open() error {
	t.file, t.err = os.Open(t.Path)
	if t.err != nil {
		return fmt.Errorf("os.Open Path %t err: %w", t.Path, t.err)
	}

	t.s = bufio.NewScanner(t.file)
	return nil
}

func (t *TxtScanner) Next() bool {
	if t.err != nil {
		return false
	}

	return t.s.Scan()
}

func (t *TxtScanner) Record() []string {
	return t.splitFunc(t.s.Text())
}

func (t *TxtScanner) Close() error {
	return t.file.Close()
}

func (t *TxtScanner) Err() error {
	return t.err
}

func FromTxt(path string, fun Action, splitFunc func(string) []string, options ...ScanOptions) error {
	return FromScanner(&TxtScanner{Path: path, splitFunc: splitFunc}, fun, options...)
}
