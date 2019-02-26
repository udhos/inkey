package inkey

import (
	"bytes"
	"io"
	//"log"
	"sync"
	//"time"
)

type Inkey struct {
	reader      io.Reader
	more        chan struct{}
	buf         bytes.Buffer
	bufMutex    sync.Mutex
	broken      error
	brokenMutex sync.RWMutex
	bufCopy     []byte
	bufLimit    int
}

func New(r io.Reader) *Inkey {
	i := &Inkey{
		reader:   r,
		more:     make(chan struct{}),
		bufCopy:  make([]byte, 100),
		bufLimit: 1000,
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

func (i *Inkey) isBroken() bool {
	return i.getBroken() != nil
}

func (i *Inkey) isFull() bool {
	i.bufMutex.Lock()
	s := i.buf.Len()
	i.bufMutex.Unlock()
	return s > i.bufLimit
}

func inputLoop(i *Inkey) {
	for {
		if !i.isBroken() && !i.isFull() {
			copy(i)
		}
		<-i.more
	}
}

func copy(i *Inkey) {

	rd, errRead := i.reader.Read(i.bufCopy)
	i.setBroken(errRead)
	//log.Printf("inputLoop: read=%d broken=%v", rd, i.getBroken())

	if rd > 0 {
		i.bufMutex.Lock()
		_, errWrite := i.buf.Write(i.bufCopy[:rd])
		//log.Printf("inputLoop: write=%d buf=%d", wr, i.buf.Len())
		i.bufMutex.Unlock()
		i.setBroken(errWrite)
	}
}

func (i *Inkey) Inkey() (byte, bool) {
	i.bufMutex.Lock()
	b, err := i.buf.ReadByte()
	i.bufMutex.Unlock()

	select {
	case i.more <- struct{}{}:
	default:
	}

	//log.Printf("Inkey: %d byte='%c' error=%v", b, b, err)

	return b, err == nil
}

func (i *Inkey) Read(buf []byte) (int, error) {

	for {
		// 1. if data in buffer, return it
		i.bufMutex.Lock()
		if i.buf.Len() > 0 {
			n, err := i.buf.Read(buf)
			i.bufMutex.Unlock()
			return n, err
		}
		i.bufMutex.Unlock()

		// 2. if error, return it
		if errBroken := i.getBroken(); errBroken != nil {
			return 0, errBroken
		}

		// 3. read more from input stream into buffer
		i.more <- struct{}{}
	}
}

func (i *Inkey) dropBackspaces() {
	buf := i.buf.Bytes()

	bufClean := make([]byte, 0, cap(buf))
	for _, b := range buf {
		switch b {
		case 8, 127:
			if len(bufClean) > 0 {
				bufClean = bufClean[:len(bufClean)-1]
			}
		default:
			bufClean = append(bufClean, b)
		}
	}
	if len(bufClean) < len(buf) {
		i.buf.Reset()
		i.buf.Write(bufClean)
	}
}

func (i *Inkey) ReadBytes(delim byte) (line []byte, err error) {

	for {
		i.bufMutex.Lock()

		// 1. remove backspaces from current buffer
		i.dropBackspaces()

		// 2. search delim in current buffer
		buf := i.buf.Bytes()
		index := bytes.IndexByte(buf, delim)

		i.bufMutex.Unlock()

		if index >= 0 {
			// found delim in current buffer
			line = make([]byte, index+1)
			_, err = i.Read(line)
			return
		}

		// 3. if error, return it
		if errBroken := i.getBroken(); errBroken != nil {
			if len(buf) > 0 {
				line = make([]byte, len(buf))
				_, err = i.Read(line)
			}
			if err == nil {
				err = errBroken
			}
			return
		}

		// 4. read more from input stream into buffer
		i.more <- struct{}{}
	}
}
