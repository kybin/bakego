package main

import (
	"bytes"
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
		encoded := hexString(data)
		reverted, err := fromHex([]byte(encoded))
		if err != nil {
			t.Fatalf("could not decode %s: %s", c.fname, err)
		}
		if !bytes.Equal(reverted, data) {
			t.Fatalf("reverted data of %s is not same as original: %s", c.fname, err)
		}
	}
}
