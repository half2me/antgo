package main

import (
	"flag"
	"os"
	file2 "github.com/half2me/antgo/driver/file"
	"encoding/gob"
	"io"
	"fmt"
)

var fileName = flag.String("file", "", "File to open")
var skip = flag.Int("skip", 0, "How many packets to skip")
var count = flag.Int("count", 0, "Read at most n packets (0 means unlimited)")

func main() {
	flag.Parse()
	file, err := os.Open(*fileName)

	if err != nil {
		panic(err.Error())
	}

	defer file.Close()

	dec := gob.NewDecoder(file)
	buf := file2.AntT{}

	// TODO: implement seek without reading to memory :/
	for i := 0; i < *skip; i++ {
		if e := dec.Decode(&buf); e != nil {
			if e == io.EOF {return} else {panic(e.Error())}
		}
	}

	if *count == 0 {
		for i := 0;;i++ {
			if e := dec.Decode(&buf); e != nil {
				if e == io.EOF {return} else {panic(e.Error())}
			}
			fmt.Printf("[%8d] %s\n", i, buf.String())
		}
	} else {
		for i := 0; i < *count ;i++ {
			if e := dec.Decode(&buf); e != nil {
				if e == io.EOF {return} else {panic(e.Error())}
			}
			fmt.Printf("[%8d] %s\n", i, buf.String())
		}
	}
}
