package main

import (
    "os"
    "fmt"
    "flag"
    "time"
    "reflect"
    "io/ioutil"
    "runtime/pprof"
    "github.com/Sereal/Sereal/Go/sereal"
)


func main() {
    file := flag.String("file", "", "sereal file to read data from")
    cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
    flag.Parse()

    fmt.Fprintf(os.Stderr, "read file: %s\n", *file)
    buf, err := ioutil.ReadFile(*file)
    if err != nil {
        panic(err)
    }

    flag.Parse()
    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            panic(err)
        }

        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    fmt.Fprintf(os.Stderr, "deserialize data\n")
    t0 := time.Now()

    dec := sereal.NewDecoder()

    var decoded interface{}
    if ok := dec.Unmarshal(buf, &decoded); ok != nil {
        panic("failed to decode file")
    }

    if *cpuprofile != "" {
        pprof.StopCPUProfile()
    }

    s := reflect.ValueOf(decoded)
    fmt.Fprintf(os.Stderr, "%d\n", s.Len())

    t1 := time.Now()
    fmt.Fprintf(os.Stderr, "Deserialization took %v to run.\n", t1.Sub(t0))
}
