package benchmark_test

import (
	"encoding/binary"
	"testing"
)

var buf []byte

func intToSlice(i int) []byte {
	if buf == nil {
		buf = make([]byte, 2048)
	}

	binary.LittleEndian.PutUint32(buf, uint32(i))
	return buf
}

func BenchmarkConvertToStringOutside(b *testing.B) {
	var str string
	m := make(map[string]interface{})

	for i := 0; i < b.N; i++ {
		str = string(intToSlice(i))
		m[str] = i
	}
}

func BenchmarkConvertToStringInside(b *testing.B) {
	var slice []byte
	m := make(map[string]interface{})

	for i := 0; i < b.N; i++ {
		slice = intToSlice(i)
		m[string(slice)] = i
	}
}
