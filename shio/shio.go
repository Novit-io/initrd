// Shared IO
package shio

import (
	"io"
	"sync"

	"novit.nc/direktil/initrd/colorio"
)

type ShIO struct {
	buf       []byte
	closed    bool
	cond      *sync.Cond
	hideInput bool
}

func New() *ShIO {
	return NewWithCap(1024)
}
func NewWithCap(capacity int) *ShIO {
	return NewWithBytes(make([]byte, 0, capacity))
}
func NewWithBytes(content []byte) *ShIO {
	return &ShIO{
		buf:  content,
		cond: sync.NewCond(&sync.Mutex{}),
	}
}

var (
	_ io.WriteCloser = New()
)

func (s *ShIO) Write(data []byte) (n int, err error) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	if s.closed {
		err = io.EOF
		return
	}

	if s.hideInput {
		s.buf = append(s.buf, colorio.Reset...)
	}

	s.buf = append(s.buf, data...)
	n = len(data)

	if s.hideInput {
		s.buf = append(s.buf, colorio.Hidden...)
	}

	s.cond.Broadcast()
	return
}

func (s *ShIO) HideInput() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	if s.closed {
		return
	}

	s.buf = append(s.buf, colorio.Hidden...)
	s.hideInput = true
}
func (s *ShIO) ShowInput() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	if s.closed {
		return
	}

	s.buf = append(s.buf, colorio.Reset...)
	s.hideInput = false
}

func (s *ShIO) Close() (err error) {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()

	s.closed = true

	s.cond.Broadcast()
	return
}

func (s *ShIO) NewReader() io.Reader {
	return &shioReader{s, 0}
}

type shioReader struct {
	in  *ShIO
	pos int
}

func (r *shioReader) Read(ba []byte) (n int, err error) {
	r.in.cond.L.Lock()
	defer r.in.cond.L.Unlock()

	for r.pos == len(r.in.buf) && !r.in.closed {
		r.in.cond.Wait()
	}

	if r.pos == len(r.in.buf) && r.in.closed {
		err = io.EOF
		return
	}

	n = copy(ba, r.in.buf[r.pos:])
	r.pos += n

	return
}
