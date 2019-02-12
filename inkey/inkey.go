package inkey

import (
	"io"
)

type Inkey struct {
	reader io.Reader
}

func New(r io.Reader) *Inkey {
	i := &Inkey{
		reader: r,
	}
	return i
}

func (i *Inkey) Inkey() (byte, bool) {
	return 'q', true
}
