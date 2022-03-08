package colorio

import "io"

var (
	Reset      = []byte("\033[0m")
	Bold       = []byte("\033[1m")
	Dim        = []byte("\033[2m")
	Underlined = []byte("\033[4m")
	Blink      = []byte("\033[5m")
	Reverse    = []byte("\033[7m")
	Hidden     = []byte("\033[8m")
)

type writer struct {
	before []byte
	after  []byte
	w      io.Writer
}

func NewWriter(color []byte, w io.Writer) io.Writer {
	return writer{
		before: color,
		after:  Reset,
		w:      w,
	}
}

func (w writer) Write(ba []byte) (n int, err error) {
	b := make([]byte, len(w.before)+len(ba)+len(w.after))
	copy(b, w.before)
	copy(b[len(w.before):], ba)
	copy(b[len(w.before)+len(ba):], w.after)

	n, err = w.w.Write(b)

	n -= len(w.before)
	if n < 0 {
		n = 0
	} else if n > len(ba) {
		n = len(ba)
	}

	return
}
