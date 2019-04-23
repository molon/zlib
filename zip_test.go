package zlib

import (
	"bytes"
	"encoding/hex"
	"io"
	"log"
	"testing"
)

func hexString(src []byte) string {
	encodedStr := hex.EncodeToString(src)
	result := ""
	index := 0
	le := len(encodedStr)
	for ; index < le; index += 8 {
		right := index + 8
		if right > le {
			right = le
		}
		result += " " + encodedStr[index:right]
	}
	if index < le {
		result += " " + encodedStr[index:]
	}
	if len(result) > 0 {
		result = result[1:]
	}

	return result
}

// go test -timeout 30s github.com/molon/zlib -run \^\(TestWriter\)\$ -v
func TestWriter(t *testing.T) {
	src := []byte("God is a girl")

	out := bytes.NewBuffer(nil)
	z, err := NewWriter(out, -15)
	if err != nil {
		log.Fatalln(err)
	}
	defer z.Close()
	if _, err := z.Write(src); err != nil {
		t.Errorf("Write failed: %v", err)
	}
	// If you dont want to use Z_SYNC_FLUSH, ignore this
	// if err := z.Flush(); err != nil {
	// 	t.Errorf("Flush failed: %v", err)
	// }
	if err := z.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}

	bts := out.Bytes()

	log.Println(hexString(bts))

	cbts := bytes.NewBuffer(bts)
	rd, err := NewReader(cbts, -15)
	if err != nil {
		t.Errorf("NewReader failed: %v", err)
	}
	defer rd.Close()

	uncompressed := bytes.NewBuffer(nil)
	_, err = io.Copy(uncompressed, rd)
	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	log.Println(uncompressed.String())
}
