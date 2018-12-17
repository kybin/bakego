// bakego bakes files' data into a go file.
package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

// readFile reads a file.
func readFile(f string) File {
	b, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	typ := String
	if !utf8.Valid(b) {
		typ = Binary
	}
	return File{fname: f, typ: typ, data: b}
}

// trimExt trims ext from file name.
func trimExt(f string) string {
	return f[:len(f)-len(filepath.Ext(f))]
}

// filePackage finds package name from existing go files.
func findPackage() string {
	files, err := filepath.Glob("*.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	pkg := ""
	for _, fname := range files {
		if strings.HasPrefix(fname, "gen_bakego") || strings.HasSuffix(trimExt(fname), "_test") {
			continue
		}
		f := readFile(fname)
		lines := bytes.Split(f.data, []byte("\n"))
		for _, line := range lines {
			line = bytes.TrimSpace(line)
			if bytes.HasPrefix(line, []byte("package ")) {
				words := bytes.Fields(line)
				if string(words[1]) != "" {
					pkg = string(words[1])
					break
				}
			}
		}
	}
	if pkg == "" {
		fmt.Fprintln(os.Stderr, "could not find package name from go file")
		os.Exit(1)
	}
	return pkg
}

// genGo generates a go file from input files.
func genGo(files []File, pkg string) {
	fd, err := os.Create("gen_bakego.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer fd.Close()
	w := bufio.NewWriter(fd)
	defer w.Flush()

	w.WriteString("// Code generated by github.com/kybin/bakego. DO NOT EDIT.\n")
	w.WriteString(fmt.Sprintf("package %s\n", pkg))
	w.WriteString(`
import (
	"bufio"
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type BakeGoFile struct {
	fname string
	enc string
	data []byte
}

type BakeGo []BakeGoFile

// Extract extracts all remebered files.
func (b BakeGo) Extract() error {
	for _, s := range b {
		fname := s.fname
		enc := s.enc
		data := s.data
		err := os.MkdirAll(filepath.Dir(fname), 0755)
		if err != nil {
			return err
		}
		f, err := os.Create(fname)
		if err != nil {
			return err
		}
		w := bufio.NewWriter(f)
		decoded := data
		if enc == "hex" {
			d, err := fromHex(data)
			if err != nil {
				return err
			}
			decoded = d
		}
		_, err = w.Write(decoded)
		if err != nil {
			return err
		}
		w.Flush()
	}
	return nil
}

// Ensure checks all the BakeGo files are exist.
// It will return nil if all files are exist or return error.
//
// It does not check the files are modified or not.
func (b BakeGo) Ensure() error {
	for _, s := range b {
		_, err := os.Stat(s.fname)
		if err != nil {
			return err
		}
	}
	return nil
}

// When developing a code, it should always be the same
// between bakego data and the actual file.
// If they are not same, it will error.
func (b BakeGo) Identical() error {
	for _, s := range b {
		src, err := ioutil.ReadFile(s.fname)
		if err != nil {
			return err
		}
		dst := s.data
		if s.enc == "hex" {
			d, err := fromHex(dst)
			if err != nil {
				return err
			}
			dst = d
		}
		if !bytes.Equal(src, dst) {
			return fmt.Errorf("bakego: src and dst is not identical: %s", s.fname)
		}
	}
	return nil
}

// fromHex are twin functions that lives both inside and outside of generated file.
// Outside one is for testing purpose.
//
// Note: if you change this function, change it's twin function too.
func fromHex(data []byte) ([]byte, error) {
	reverted := make([]byte, 0, len(data))
	lines := bytes.Split(data, []byte("\n"))
	for _, src := range lines {
		dst := make([]byte, hex.DecodedLen(len(src)))
		_, err := hex.Decode(dst, src)
		if err != nil {
			return nil, err
		}
		reverted = append(reverted, dst...)
	}
	return reverted, nil
}

var bakego BakeGo = make([]BakeGoFile, 0)

func init() {
`)
	sort.Slice(files, func(i, j int) bool {
		return files[i].fname < files[j].fname
	})
	for _, f := range files {
		if f.typ == Unknown {
			continue
		}
		enc := ""
		if f.typ == Binary {
			enc = "hex"
		}
		w.WriteString(fmt.Sprintf("\tbakego = append(bakego, BakeGoFile{\"%s\", \"%s\", []byte(`", f.fname, enc))
		if f.typ == String {
			// raw string cannot handle ` (quote), make them separate strings
			data := bytes.Replace(f.data, []byte("`"), []byte("` + \"`\" + `"), -1)
			w.Write(data)
		} else if f.typ == Binary {
			w.Write(toHex(f.data))
		}
		w.WriteString("`)})\n")
	}
	w.WriteString("}\n")
}

func toHex(data []byte) []byte {
	remain := data
	hex := []byte{}
	cut := 64
	for {
		if len(remain) < cut {
			cut = len(remain)
		}
		c := remain[:cut]
		hex = append(hex, []byte(fmt.Sprintf("%x\n", c))...)
		remain = remain[cut:]
		if len(remain) == 0 {
			break
		}
	}
	return hex
}

// fromHex are twin functions that lives both inside and outside of generated file.
// Outside one is for testing purpose.
//
// Note: if you change this function, change it's twin function too.
func fromHex(data []byte) ([]byte, error) {
	reverted := make([]byte, 0, len(data))
	lines := bytes.Split(data, []byte("\n"))
	for _, src := range lines {
		dst := make([]byte, hex.DecodedLen(len(src)))
		_, err := hex.Decode(dst, src)
		if err != nil {
			return nil, err
		}
		reverted = append(reverted, dst...)
	}
	return reverted, nil
}

func genGoTest() {
	fd, err := os.Create("gen_bakego_test.go")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer fd.Close()
	w := bufio.NewWriter(fd)
	defer w.Flush()
	w.WriteString(`package main

import "testing"

func TestBakeGo(t *testing.T) {
	err := bakego.Identical()
	if err != nil {
		t.Fatal(err)
	}
}`)
}

type FileType int

const (
	Unknown = FileType(iota)
	String
	Binary
)

type File struct {
	fname string
	typ   FileType
	data  []byte
}

func main() {
	var (
		recursive bool
		dir       bool
	)
	flag.BoolVar(&dir, "d", false, "add files inside of the directory")
	flag.BoolVar(&recursive, "r", false, "add files inside of the directory recursively")
	flag.Parse()
	elems := flag.Args()
	if len(elems) == 0 {
		fmt.Fprintln(os.Stderr, "ex) bakego [args...] [file|dir...]")
		os.Exit(1)
	}
	files := make([]File, 0)
	for _, el := range elems {
		fi, err := os.Stat(el)
		if err != nil {
			fmt.Fprint(os.Stderr, err)
			os.Exit(1)
		}
		if fi.IsDir() {
			if !dir && !recursive {
				fmt.Fprintln(os.Stderr, "if you want to add a directory, use -d or -r flag")
				os.Exit(1)
			}
			d := len(filepath.SplitList(el))
			filepath.Walk(el, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if recursive {
						return nil
					}
					wd := len(filepath.SplitList(path))
					if wd-d <= 1 {
						return nil
					}
					return filepath.SkipDir
				}
				files = append(files, readFile(path))
				return nil
			})
		} else {
			files = append(files, readFile(el))
		}
	}
	pkg := findPackage()
	genGo(files, pkg)
	genGoTest()
}
