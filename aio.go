package log

import (
	"io"
	"runtime"
	"sync/atomic"
)

type aio struct {
	b  []byte
	w  io.Writer
	ch chan struct{}
}

// AIO implements buffering asynchronous Writer for an io.Writer object.
// Which can reduce the latency spike of api occurrence by disk/system latency.
// If an error occurs writing to a Writer, no more data will be
// accepted and all subsequent writes, and Flush, will return the error.
// After all data has been written, the client should call the
// Flush method to guarantee all data has been forwarded to
// the underlying io.Writer.
type AIO struct {
	fault   *atomic.Value
	buf     []byte
	n, size int
	w       io.Writer
	ch      chan *aio
	shared  chan []byte
}

// NewAIO returns a new Writer whose buffer has at least the specified
// size. If the argument io.Writer is already a Writer with large enough
// size, it returns the underlying Writer.
func NewAIO(w io.Writer, size int) *AIO {
	a := &AIO{
		fault:  &atomic.Value{},
		buf:    make([]byte, size),
		size:   size,
		w:      w,
		ch:     make(chan *aio, 128),
		shared: make(chan []byte, 128),
	}
	go loop(a.ch, a.shared, a.fault)
	runtime.SetFinalizer(a, func(a *AIO) { close(a.ch) })
	return a
}

func loop(reqch chan *aio, shared chan []byte, fault *atomic.Value) {
	for req := range reqch {
		if len(req.b) != 0 && req.w != nil {
			n, err := req.w.Write(req.b)
			if n < len(req.b) && err == nil {
				err = io.ErrShortWrite
			}
			if err == nil {
				select {
				case shared <- req.b:
				default:
				}
			} else {
				fault.Store(struct{ error }{err})
			}
		}
		if req.ch != nil {
			close(req.ch)
		}
	}
}

// Reset discards any unflushed buffered data, clears any error, and
// resets b to write its output to w.
func (a *AIO) Reset(w io.Writer) {
	a.fault.Store(struct{ error }{nil})
	a.n = 0
	a.w = w
}

func (a *AIO) haserror() error {
	err, _ := a.fault.Load().(struct{ error })
	return err.error
}

func (a *AIO) free() []byte {
	select {
	case b := <-a.shared:
		return b[:cap(b)]
	default:
		return make([]byte, a.size)
	}
}

// Flush writes any buffered data to the underlying io.Writer.
func (a *AIO) Flush() error {
	if e := a.haserror(); e != nil {
		return e
	}
	aio := &aio{ch: make(chan struct{})}
	if a.n != 0 {
		aio.w = a.w
		aio.b = a.buf[:a.n]
		a.buf = a.free()
		a.n = 0
	}
	a.ch <- aio
	<-aio.ch
	return a.haserror()
}

func (a *AIO) flush() {
	aio := &aio{
		w: a.w,
		b: a.buf[:a.n],
	}
	a.buf = a.free()
	a.n = 0
	a.ch <- aio
}

// Available returns how many bytes are unused in the buffer.
func (a *AIO) Available() int { return len(a.buf) - a.n }

// Buffered returns the number of bytes that have been written into the current buffer.
func (a *AIO) Buffered() int { return a.n }

// Write writes the contents of p into the buffer.
// It returns the number of bytes written.
// If nn < len(p), it also returns an error explaining
// why the write is short.
func (a *AIO) Write(p []byte) (nn int, err error) {
	for len(p) > a.Available() && a.haserror() == nil {
		n := copy(a.buf[a.n:], p)
		a.n += n
		a.flush()
		nn += n
		p = p[n:]
	}
	if e := a.haserror(); e != nil {
		return nn, e
	}
	n := copy(a.buf[a.n:], p)
	a.n += n
	nn += n
	return nn, nil
}
