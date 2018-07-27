package log

import (
	"bytes"
	"io"
	"io/ioutil"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAIOAtGC(t *testing.T) {
	i := 0
	aa := NewAIO(ioutil.Discard, 1024)
	runtime.SetFinalizer(aa, nil)
	runtime.SetFinalizer(aa, func(a *AIO) { close(a.ch); i = 1 })
	runtime.KeepAlive(aa)
	aa = nil
	runtime.GC()
	runtime.GC()
	assert.Equal(t, 1, i)
}

func TestAIOReset(t *testing.T) {
	NewAIO(ioutil.Discard, 1024).Reset(ioutil.Discard)
}

func TestAIGeneric(t *testing.T) {
	var (
		assert = assert.New(t)
		w0     = bytes.NewBuffer(nil)
		aio    = NewAIO(w0, 128)
	)

	assert.Equal(128, aio.Available())
	n, err := aio.Write([]byte("abcdef"))
	assert.Equal(6, n)
	assert.Equal(nil, err)
	assert.Equal(128-6, aio.Available())
	n, err = aio.Write([]byte("abcdef"))
	assert.Equal(6, n)
	assert.Equal(nil, err)
	assert.Equal(128-12, aio.Available())
	assert.Equal(12, aio.Buffered())
	assert.Equal(nil, aio.Flush())
	assert.Equal("abcdefabcdef", w0.String())
	assert.Equal(128, aio.Available())

	n, err = aio.Write([]byte("abcdef"))
	assert.Equal(6, n)
	assert.Equal(nil, err)
	assert.Equal(128-6, aio.Available())
	n, err = aio.Write(bytes.Repeat([]byte("abcd"), 128))
	assert.Equal(128*4, n)
	assert.Equal(nil, err)

	aio.Reset(&faultbuf{})
	assert.Equal(128, aio.Available())
	aio.Write([]byte("abcdef"))
	assert.Equal(io.ErrClosedPipe, aio.Flush())
}

type faultbuf struct{}

func (b *faultbuf) Write(p []byte) (int, error) {
	return 0, io.ErrClosedPipe
}
