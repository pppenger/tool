package main

import (
	"io"
	"sync"
)

type LockedWriter struct {
	m      sync.Mutex
	Writer io.Writer
}

func (lw *LockedWriter) Write(b []byte) (n int, err error) {
	lw.m.Lock()
	defer lw.m.Unlock()
	return lw.Writer.Write(b)
}

func NewLockedWriter(w io.Writer) io.Writer {
	return &LockedWriter{
		Writer: w,
	}
}
