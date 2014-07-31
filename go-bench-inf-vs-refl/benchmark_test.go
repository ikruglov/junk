package benchmark_test 

import (
    "reflect"
    "testing"
)

func BenchmarkAssignInf(b *testing.B) {
    arr := make([]interface{}, b.N)
    for i := 0; i < b.N; i++ {
        arr[i] = "string"
    }
}

func BenchmarkAssignInfViaReflection(b *testing.B) {
    arr := make([]interface{}, b.N)
    rv := reflect.ValueOf(arr)

    idxs := make([]reflect.Value, b.N)
    for i := 0; i < b.N; i++ {
        idxs[i] = rv.Index(i)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idxs[i].Set(reflect.ValueOf("string"))
    }
}

func BenchmarkAssignString(b *testing.B) {
    arr := make([]string, b.N)
    for i := 0; i < b.N; i++ {
        arr[i] = "string"
    }
}

func BenchmarkAssignStringViaRelfection(b *testing.B) {
    arr := make([]string, b.N)
    rv := reflect.ValueOf(arr)

    idxs := make([]reflect.Value, b.N)
    for i := 0; i < b.N; i++ {
        idxs[i] = rv.Index(i)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idxs[i].SetString("string")
    }
}
