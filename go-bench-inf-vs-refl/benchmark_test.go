package benchmark_test

import (
    _ "fmt"
    "strconv"
    "reflect"
    "testing"
)

func BenchmarkAssignArrayInterface1(b *testing.B) {
    arr := make([]interface{}, b.N)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        arr[i] = "string" + strconv.Itoa(i)
    }
}

func BenchmarkAssignArrayInterface2(b *testing.B) {
    arr := make([]interface{}, b.N)
    rv := reflect.ValueOf(arr)

    idxs := make([]reflect.Value, b.N)
    for i := 0; i < b.N; i++ {
        idxs[i] = rv.Index(i)
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        idxs[i].Set(reflect.ValueOf("string" + strconv.Itoa(i)))
    }
}

func BenchmarkAssignArrayInterface3(b *testing.B) {
    arr := make([]interface{}, b.N)
    rv := reflect.ValueOf(arr)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rv.Index(i).Set(reflect.ValueOf("string" + strconv.Itoa(i)))
    }
}

func BenchmarkAssignArrayString1(b *testing.B) {
    arr := make([]string, b.N)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        arr[i] = "string" + strconv.Itoa(i)
    }
}

func BenchmarkAssignArrayString2(b *testing.B) {
    arr := make([]string, b.N)
    rv := reflect.ValueOf(arr)

    idxs := make([]reflect.Value, b.N)
    for i := 0; i < b.N; i++ {
        idxs[i] = rv.Index(i)
    }

    b.ResetTimer()
    for i := 0; i < len(arr); i++ {
        idxs[i].SetString("string" + strconv.Itoa(i))
    }
}

func BenchmarkAssignArrayString3(b *testing.B) {
    arr := make([]string, b.N)
    rv := reflect.ValueOf(arr)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rv.Index(i).SetString("string" + strconv.Itoa(i))
    }
}

func BenchmarkAssignArrayString4(b *testing.B) {
    arr := make([]string, b.N)
    rv := reflect.ValueOf(arr)

    idxs := make([]reflect.Value, b.N)
    for i := 0; i < b.N; i++ {
        idxs[i] = rv.Index(i)
    }

    b.ResetTimer()
    for i := 0; i < len(arr); i++ {
        idxs[i].Set(reflect.ValueOf("string" + strconv.Itoa(i)))
    }
}

func BenchmarkAssignArrayString5(b *testing.B) {
    arr := make([]string, b.N)
    rv := reflect.ValueOf(arr)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        rv.Index(i).Set(reflect.ValueOf("string" + strconv.Itoa(i)))
    }
}

func BenchmarkAssignHashInterface1(b *testing.B) {
    hash := make(map[string]interface{}, b.N)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        str := "string" + strconv.Itoa(i)
        hash[str] = str
    }
}

func BenchmarkAssignHashInterface2(b *testing.B) {
    hash := make(map[string]interface{}, b.N)
    rv := reflect.ValueOf(hash)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        str := reflect.ValueOf("string" + strconv.Itoa(i))
        rv.SetMapIndex(str, str)
    }
}
