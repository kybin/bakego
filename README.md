# bakego

bakego bakes files into a go file.

## Install

Install Go, then

```
go get github.com/kybin/bakego
```

## Usage Example

bakego created for github.com/kybin/whisky at first.
It has template files in tmpl directory like any usual server program.

In this case you can,

```
$ cd whisky
$ bakego -d tmpl
```

It will generate gen_bakego.go.
The generated code has a global variable named bakego, itself has two methods Extract and Ensure.
So you can add code yourself like,

```
// add init flag to program, then in somewhere of main function

if initFlag {
	bakego.Extract()
} else {
	bakego.Ensure()
}
```

After `go install`, `whisky -init` will create tmpl directory in the run directory.
