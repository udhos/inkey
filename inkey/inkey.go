package inkey

import (
	"bytes"
	"io"
	"log"
	"sync"
	"time"
)

type Inkey struct {
	reader io.Reader
	//more        chan struct{}
	buf         bytes.Buffer
	bufMutex    sync.Mutex
	broken      error
	brokenMutex sync.RWMutex
}

func New(r io.Reader) *Inkey {
	i := &Inkey{
		reader: r,
		//more:   make(chan struct{}),
	}
	go inputLoop(i)
	return i
}

func (i *Inkey) getBroken() error {
	i.brokenMutex.RLock()
	err := i.broken
	i.brokenMutex.RUnlock()
	return err
}

func (i *Inkey) setBroken(err error) {
	i.brokenMutex.Lock()
	if i.broken == nil {
		i.broken = err
	}
	i.brokenMutex.Unlock()
}

func full(i *Inkey) bool {
	i.bufMutex.Lock()
	s := i.buf.Len()
	i.bufMutex.Unlock()
	return s > 100
}

func inputLoop(i *Inkey) {
	for {
		if i.getBroken() == nil {
			copy(i)
		}
		time.Sleep(time.Second) // ugh
	}
}

func copy(i *Inkey) {

	buf := make([]byte, 10)

	if full(i) {
		return
	}

	rd, errRead := i.reader.Read(buf)
	i.setBroken(errRead)
	log.Printf("inputLoop: read=%d broken=%v", rd, i.getBroken())

	if rd > 0 {
		i.bufMutex.Lock()
		wr, errWrite := i.buf.Write(buf[:rd])
		log.Printf("inputLoop: write=%d buf=%d", wr, i.buf.Len())
		i.bufMutex.Unlock()
		i.setBroken(errWrite)
	}
}

func (i *Inkey) Inkey() (byte, bool) {
	i.bufMutex.Lock()
	b, err := i.buf.ReadByte()
	i.bufMutex.Unlock()

	log.Printf("Inkey: %d byte='%c' error=%v", b, b, err)

	return b, err == nil
}
