package main

import "fmt"

func main() {
	h := NewHash()
	ints := []uint32{32, 44, 36, 9, 25, 5, 14, 18, 10, 30, 31, 35, 7, 11}
	for _, i := range ints {
		h.Insert(uint32(i), 0)
		h.Dump()
	}

	for _, i := range ints {
		if !h.Has(uint32(i)) {
			fmt.Println("don't have", i)
			h.Dump()
			break
		}
	}

	// k := 8
	// for i := 0; i < k; i++ {
	// h.Insert(uint32(i), 0)
	// h.Dump()
	// }

	// for i := 0; i < k; i++ {
	// if !h.Has(uint32(i)) {
	// fmt.Println("don't have", i)
	// h.Dump()
	// break
	// }
	// }
}
