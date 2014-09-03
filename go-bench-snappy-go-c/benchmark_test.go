package benchmark_test

import (
	"testing"
    "io/ioutil"
    "github.com/dgryski/go-csnappy"
    "code.google.com/p/snappy-go/snappy"
)

func BenchmarkCSnappy(b *testing.B) {
    src, err := ioutil.ReadFile("data")
    if err != nil {
        b.Fatal(err)
    }

    var dst []byte
    for i := 0; i < b.N; i++ {
        _, err := csnappy.Encode(dst, src)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkGoSnappy(b *testing.B) {
    src, err := ioutil.ReadFile("data")
    if err != nil {
        b.Fatal(err)
    }

    var dst []byte
    for i := 0; i < b.N; i++ {
        _, err := snappy.Encode(dst, src)
        if err != nil {
            b.Fatal(err)
        }
    }
}
