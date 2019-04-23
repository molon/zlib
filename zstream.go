// Copyright 2015, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package zlib

// See http://www.zlib.net/zlib_how.html for more information on this

/*
#cgo CFLAGS: -Werror=implicit
#cgo pkg-config: zlib

#include "zlib.h"

// inflateInit2 is a macro, so using a wrapper function
int zstream_inflate_init(char *strm, int windowBits) {
  ((z_stream*)strm)->zalloc = Z_NULL;
  ((z_stream*)strm)->zfree = Z_NULL;
  ((z_stream*)strm)->opaque = Z_NULL;
  ((z_stream*)strm)->avail_in = 0;
  ((z_stream*)strm)->next_in = Z_NULL;
  return inflateInit2((z_stream*)strm, windowBits);
}

// deflateInit2 is a macro, so using a wrapper function
int zstream_deflate_init(char *strm, int level, int method, int windowBits, int memLevel, int strategy) {
  ((z_stream*)strm)->zalloc = Z_NULL;
  ((z_stream*)strm)->zfree = Z_NULL;
  ((z_stream*)strm)->opaque = Z_NULL;
  return deflateInit2((z_stream*)strm, level, method, windowBits, memLevel, strategy);
}

unsigned int zstream_avail_in(char *strm) {
  return ((z_stream*)strm)->avail_in;
}

unsigned int zstream_avail_out(char *strm) {
  return ((z_stream*)strm)->avail_out;
}

char* zstream_msg(char *strm) {
  return ((z_stream*)strm)->msg;
}

void zstream_set_in_buf(char *strm, void *buf, unsigned int len) {
  ((z_stream*)strm)->next_in = (Bytef*)buf;
  ((z_stream*)strm)->avail_in = len;
}

void zstream_set_out_buf(char *strm, void *buf, unsigned int len) {
  ((z_stream*)strm)->next_out = (Bytef*)buf;
  ((z_stream*)strm)->avail_out = len;
}

int zstream_inflate(char *strm, int flag) {
  return inflate((z_stream*)strm, flag);
}

int zstream_deflate(char *strm, int flag) {
  return deflate((z_stream*)strm, flag);
}

void zstream_inflate_end(char *strm) {
  inflateEnd((z_stream*)strm);
}

void zstream_deflate_end(char *strm) {
  deflateEnd((z_stream*)strm);
}
*/
import "C"

import (
	"fmt"
	"unsafe"
)

const (
	// Version
	ZLIB_Version = C.ZLIB_VERSION

	// Allowed flush values
	Z_NO_FLUSH      = C.Z_NO_FLUSH
	Z_PARTIAL_FLUSH = 1
	Z_SYNC_FLUSH    = 2
	Z_FULL_FLUSH    = 3
	Z_FINISH        = 4
	Z_BLOCK         = 5
	Z_TREES         = 6

	// Return codes
	Z_OK            = 0
	Z_STREAM_END    = 1
	Z_NEED_DICT     = 2
	Z_ERRNO         = -1
	Z_STREAM_ERROR  = -2
	Z_DATA_ERROR    = -3
	Z_MEM_ERROR     = -4
	Z_BUF_ERROR     = -5
	Z_VERSION_ERROR = -6

	// compression levels
	Z_NO_COMPRESSION      = 0
	Z_BEST_SPEED          = 1
	Z_BEST_COMPRESSION    = 9
	Z_DEFAULT_COMPRESSION = -1

	// method
	Z_DEFLATED = 8

	// strategy
	Z_DEFAULT_STRATEGY = 0

	// our default buffer size
	// most go io functions use 32KB as buffer size, so 32KB
	// works well here for compressed data buffer
	DEFAULT_COMPRESSED_BUFFER_SIZE = 32 * 1024
)

// z_stream is a buffer that's big enough to fit a C.z_stream.
// This lets us allocate a C.z_stream within Go, while keeping the contents
// opaque to the Go GC. Otherwise, the GC would look inside and complain that
// the pointers are invalid, since they point to objects allocated by C code.
type zstream [unsafe.Sizeof(C.z_stream{})]C.char

func (strm *zstream) inflateInit(windowBits int) error {
	result := C.zstream_inflate_init(&strm[0], C.int(windowBits))
	if result != Z_OK {
		return fmt.Errorf("zlib: failed to initialize inflate (%v): %v", result, strm.msg())
	}
	return nil
}

func (strm *zstream) deflateInit(level int, method int, windowBits int, memLevel int, strategy int) error {
	result := C.zstream_deflate_init(&strm[0], C.int(level), C.int(method), C.int(windowBits), C.int(memLevel), C.int(strategy))
	if result != Z_OK {
		return fmt.Errorf("zlib: failed to initialize deflate (%v): %v", result, strm.msg())
	}
	return nil
}

func (strm *zstream) inflateEnd() {
	C.zstream_inflate_end(&strm[0])
}

func (strm *zstream) deflateEnd() {
	C.zstream_deflate_end(&strm[0])
}

func (strm *zstream) availIn() int {
	return int(C.zstream_avail_in(&strm[0]))
}

func (strm *zstream) availOut() int {
	return int(C.zstream_avail_out(&strm[0]))
}

func (strm *zstream) msg() string {
	return C.GoString(C.zstream_msg(&strm[0]))
}

func (strm *zstream) setInBuf(buf []byte, size int) {
	if buf == nil {
		C.zstream_set_in_buf(&strm[0], nil, C.uint(size))
	} else {
		C.zstream_set_in_buf(&strm[0], unsafe.Pointer(&buf[0]), C.uint(size))
	}
}

func (strm *zstream) setOutBuf(buf []byte, size int) {
	if buf == nil {
		C.zstream_set_out_buf(&strm[0], nil, C.uint(size))
	} else {
		C.zstream_set_out_buf(&strm[0], unsafe.Pointer(&buf[0]), C.uint(size))
	}
}

func (strm *zstream) inflate(flag int) (int, error) {
	ret := C.zstream_inflate(&strm[0], C.int(flag))
	switch ret {
	case Z_NEED_DICT:
		ret = Z_DATA_ERROR
		fallthrough
	case Z_DATA_ERROR, Z_MEM_ERROR:
		return int(ret), fmt.Errorf("zlib: failed to inflate (%v): %v", ret, strm.msg())
	}
	return int(ret), nil
}

func (strm *zstream) deflate(flag int) {
	ret := C.zstream_deflate(&strm[0], C.int(flag))
	if ret == Z_STREAM_ERROR {
		// all the other error cases are normal,
		// and this should never happen
		panic(fmt.Errorf("zlib: Unexpected error (1)"))
	}
}
