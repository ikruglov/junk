package main

import (
	"flag"
	"fmt"
	"github.com/Sereal/Sereal/Go/sereal"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"time"
)

func main() {
	pcnt := flag.Int("n", 1, "workers")
	folder := flag.String("folder", ".", "folder to read *.srl files from")
	flag.Parse()

	files, _ := filepath.Glob(*folder + "/*.srl")
	sort.Strings(files)

	if len(files) == 0 {
		panic("no sereal files read")
	}

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

	t0 := time.Now()

	ln := len(data)
	ch := make(chan []byte)
	fmt.Fprintf(os.Stderr, "workers: %d total size: %d\n", *pcnt, ln)

	runtime.GOMAXPROCS(20)

	for i := 0; i < *pcnt; i++ {
		from := i * (ln / *pcnt)
		to := (i + 1) * (ln / *pcnt)
		if i == *pcnt-1 {
			to = len(data)
		}

		go func(slice []interface{}, from, to int) {
			fmt.Fprintf(os.Stderr, "start worker: %d -> %d (len %d)\n", from, to, len(slice))
			tg0 := time.Now()
			enc := sereal.NewEncoderV3()
			encoded, err := enc.Marshal(slice)
			if err != nil {
				panic("failed to marshal data")
			}

			tg1 := time.Now()
			fmt.Fprintf(os.Stderr, "finished encoding %d -> %d (%v)\n", from, to, tg1.Sub(tg0))
			ch <- encoded
		}(data[from:to], from, to)
	}

	mrg := sereal.NewMerger()
	mrg.TopLevelElement = sereal.TopLevelArray
	mrg.KeepFlat = true

	for i := 0; i < *pcnt; i++ {
		if _, err := mrg.Append(<-ch); err != nil {
			panic("failed to append to merger")
		}
	}

	merged, err := mrg.Finish()
	if err != nil {
		panic("failed to get result from merger")
	}

	t1 := time.Now()
	fmt.Printf("%s", merged)
	fmt.Fprintf(os.Stderr, "The call took %v to run.\n", t1.Sub(t0))
}
