package main

import (
	"flag"
	"fmt"
	"github.com/Sereal/Sereal/Go/sereal"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

import _ "net/http/pprof"

func main() {
	folder := flag.String("folder", ".", "folder to read *.srl files from")
	flag.Parse()

	files, _ := filepath.Glob(*folder + "/*.srl")
	sort.Strings(files)

	if len(files) == 0 {
		panic("no sereal files read")
	}

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	var data []interface{}
	dec := sereal.NewDecoder()

	fmt.Fprintf(os.Stderr, "read and deserialize data\n")
	for _, file := range files {
		buf, ok := ioutil.ReadFile(file)
		if ok != nil {
			panic("failed to read file: " + file)
		}

		var decoded interface{}
		if ok := dec.Unmarshal(buf, &decoded); ok != nil {
			panic("failed to decode file: " + file)
		}

		data = append(data, decoded)
	}

	//for {
	t0 := time.Now()
	fmt.Fprintf(os.Stderr, "serialize data\n")

	enc := sereal.NewEncoderV3()
	encoded, err := enc.Marshal(data)
	if err != nil {
		panic("failed to marshal data")
	}

	t1 := time.Now()
	fmt.Printf("%s", encoded)
	fmt.Fprintf(os.Stderr, "The call took %v to run.\n", t1.Sub(t0))
	//}
}
