// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib

import (
	"fmt"
	"io"
)

// err starts out as nil
// we will call deflateEnd when we set err to a value:
// - whatever error is returned by the underlying writer
// - io.EOF if Close was called
type Writer struct {
	w    io.Writer
	out  []byte
	strm zstream
	err  error
}

func NewWriter(w io.Writer, windowBits int) (*Writer, error) {
	return NewWriterLevel(w, Z_DEFAULT_COMPRESSION, windowBits)
}

func NewWriterLevel(w io.Writer, level int, windowBits int) (*Writer, error) {
	return NewWriterLevelBuffer(w, DEFAULT_COMPRESSED_BUFFER_SIZE, level, Z_DEFLATED, windowBits, 8, Z_DEFAULT_STRATEGY)
}

func NewWriterLevelBuffer(w io.Writer, bufferSize int, level int, method int, windowBits int, memLevel int, strategy int) (*Writer, error) {
	z := &Writer{w: w, out: make([]byte, bufferSize)}
	if err := z.strm.deflateInit(level, method, windowBits, memLevel, strategy); err != nil {
		return nil, err
	}
	return z, nil
}

// this is the main function: it advances the write with either
// new data or something else to do, like a flush
func (z *Writer) write(p []byte, flush int) int {
	if len(p) == 0 {
		z.strm.setInBuf(nil, 0)
	} else {
		z.strm.setInBuf(p, len(p))
	}
	// we loop until we don't get a full output buffer
	// each loop completely writes the output buffer to the underlying
	// writer
	for {
		// deflate one buffer
		z.strm.setOutBuf(z.out, len(z.out))
		z.strm.deflate(flush)

		// write everything
		from := 0
		have := len(z.out) - int(z.strm.availOut())
		for have > 0 {
			var n int
			n, z.err = z.w.Write(z.out[from:have])
			if z.err != nil {
				z.strm.deflateEnd()
				return 0
			}
			from += n
			have -= n
		}

		// we stop trying if we get a partial response
		if z.strm.availOut() != 0 {
			break
		}
	}
	// the library guarantees this
	if z.strm.availIn() != 0 {
		panic(fmt.Errorf("zlib: Unexpected error (2)"))
	}
	return len(p)
}

func (z *Writer) Write(p []byte) (n int, err error) {
	if z.err != nil {
		return 0, z.err
	}
	n = z.write(p, Z_NO_FLUSH)
	return n, z.err
}

func (z *Writer) Flush() error {
	if z.err != nil {
		return z.err
	}
	z.write(nil, Z_SYNC_FLUSH)
	return z.err
}

// Calling Close does not close the wrapped io.Writer originally
// passed to NewWriterX.
func (z *Writer) Close() error {
	if z.err != nil {
		return z.err
	}
	z.write(nil, Z_FINISH)
	if z.err != nil {
		return z.err
	}
	z.strm.deflateEnd()
	z.err = io.EOF
	return nil
}
