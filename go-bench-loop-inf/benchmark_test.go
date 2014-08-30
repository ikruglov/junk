package benchmark_test

import (
    "testing"
    "strconv"
)

func BenchmarkDeclareInterfaceInside(b *testing.B) {
    hash := make(map[string]interface{}, b.N)

    for i := 0; i < b.N; i++ {
        var iface interface{}
        iface = i
        hash[strconv.Itoa(i)] = iface
    }
}

func BenchmarkDeclareInterfaceOutside(b *testing.B) {
    hash := make(map[string]interface{}, b.N)
    var iface interface{}

    for i := 0; i < b.N; i++ {
        iface = i
        hash[strconv.Itoa(i)] = iface
    }
}
