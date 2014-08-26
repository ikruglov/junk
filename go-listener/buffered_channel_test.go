package main

import (
	"testing"
)

func BenchmarkBufferedChannel(b *testing.B) {
	data := make([]byte, 10)
	done := make(chan int)
	in, out := MakeBufferedChannel()

	go func(ch <-chan []byte, done chan int) {
		cnt := 0
		for i := 0; i < b.N; i++ {
			<-ch
			cnt++
		}

		done <- cnt
	}(out, done)

	for i := 0; i < b.N; i++ {
		in <- data
	}

	cnt := <-done
	if b.N != cnt {
		b.Fatal("go routine returned wrong number", cnt, b.N)
	}
}
