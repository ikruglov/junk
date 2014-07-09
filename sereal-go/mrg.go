package main

import (
	"flag"
	"fmt"
	"github.com/Sereal/Sereal/Go/sereal"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type srl_file struct {
	b    []byte
	name string
}

func main() {
	folder := flag.String("folder", ".", "folder to read *.srl files from")
	flag.Parse()

	files, _ := filepath.Glob(*folder + "/*.srl")
	sort.Strings(files)

	fmt.Fprintf(os.Stderr, "read data\n")

	var data []srl_file
	for _, file := range files {
		buf, ok := ioutil.ReadFile(file)
		if ok != nil {
			panic("failed to read file: " + file)
		}

		data = append(data, srl_file{buf, file})
	}

	fmt.Fprintf(os.Stderr, "merge data\n")
	t0 := time.Now()

	m := sereal.NewMerger()
	m.DedupeStrings = true
	m.KeepFlat = true

	//m.Compression = sereal.SnappyCompressor{Incremental: true}
	//m.Compression = sereal.ZlibCompressor{}

	for _, srl := range data {
		_, err := m.Append(srl.b)
		if err != nil {
			panic(err)
		}
	}

	menc, _ := m.Finish()
	fmt.Printf("%s", menc)

	t1 := time.Now()
	fmt.Fprintf(os.Stderr, "Merging took %v to run.\n", t1.Sub(t0))
}
