// +build ignore
package main

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"unicode/utf8"
)

func dieIf(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	data, err := ioutil.ReadFile("a.png")
	dieIf(err)
	ok := utf8.Valid(data)
	if !ok {
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
			dieIf(err)
			revertedData = append(revertedData, dst...)
		}
		err := ioutil.WriteFile("b.png", revertedData, 0644)
		dieIf(err)
	}
}
