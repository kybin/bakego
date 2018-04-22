package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"testing"
)

func dieIf(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func TestHexEncoding(t *testing.T) {
	cases := []struct {
		fname string
	}{
		{"testdata/kybin.png"},
	}
	for _, c := range cases {
		data, err := ioutil.ReadFile(c.fname)
		if err != nil {
			t.Fatalf("could not read %s: %s", c.fname, err)
		}
		remain := data
		stream := []byte{}
		cut := 64
		for {
			if len(remain) < cut {
				cut = len(remain)
			}
			c := remain[:cut]
			stream = append(stream, []byte(fmt.Sprintf("%x\n", c))...)
			remain = remain[cut:]
			if len(remain) == 0 {
				break
			}
		}
		revertedData := make([]byte, 0, len(data))
		streams := bytes.Split(stream, []byte("\n"))
		for _, src := range streams {
			dst := make([]byte, hex.DecodedLen(len(src)))
			_, err := hex.Decode(dst, src)
			if err != nil {
				t.Fatalf("could not decode %s: %s", c.fname, err)
			}
			revertedData = append(revertedData, dst...)
		}
		if !bytes.Equal(revertedData, data) {
			t.Fatalf("reverted data of %s is not same as original: %s", c.fname, err)
		}
	}
}
