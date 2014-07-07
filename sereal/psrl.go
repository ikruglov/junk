package main

import (
    "os"
    "fmt"
    "time"
    "sort"
    "flag"
    "io/ioutil"
    "runtime/pprof"
    "path/filepath"
    "github.com/Sereal/Sereal/Go/sereal"
)

func main() {
    folder := flag.String("folder", ".", "folder to read *.srl files from")
    cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
    memprofile := flag.String("memprofile", "", "write memory profile to this file")
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

    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            panic(err)
        }

        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    fmt.Fprintf(os.Stderr, "serialize data\n")
    t0 := time.Now()

    enc := sereal.NewEncoderV3()
    encoded, err := enc.Marshal(data)
    if err != nil {
        fmt.Println("failed to marshal data")
        return
    }

    t1 := time.Now()

    if *memprofile != "" {
        f, err := os.Create(*memprofile)
        if err != nil {
            panic(err)
        }
        pprof.WriteHeapProfile(f)
        f.Close()
        return
    }

    if *cpuprofile != "" {
        pprof.StopCPUProfile()
    }

    fmt.Printf("%s", encoded)
    fmt.Fprintf(os.Stderr, "The call took %v to run.\n", t1.Sub(t0))
}
