package colorio

import (
	"bytes"
	"fmt"
	"testing"
)

func TestWriter(t *testing.T) {
	buf := new(bytes.Buffer)

	w := NewWriter(Bold, buf)

	buf.WriteByte('{')
	fmt.Fprintln(w, "hello")
	buf.WriteByte('}')

	if s, exp := buf.String(), "{"+string(Bold)+"hello\n"+string(Reset)+"}"; s != exp {
		t.Errorf("%q != %q", s, exp)
	}
}
