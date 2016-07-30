package main

import "fmt"

const RecordsPerBucket = 4

type Record struct {
	key   uint32
	hash  uint32
	value uint32
}

type Hash struct {
	buckets       [][]*Record
	records       uint32
	bits          uint32
	bucketToSplit uint32
	dosplit       bool
}

func NewHash() *Hash {
	return &Hash{buckets: make([][]*Record, 2), bits: 1}
}

func (h *Hash) Insert(key uint32, value uint32) {
	hash := key
	fmt.Printf("Insert: %d %b\n", hash, hash)
	bucket := h.bucket(hash)
	fmt.Println("Insert: bucket", bucket)

	h.records++
	h.buckets[bucket] = append(h.buckets[bucket], &Record{hash, hash, value})
	if len(h.buckets[bucket]) > RecordsPerBucket || h.bucketToSplit > 0 {
		h.split()
	}
}

func (h *Hash) Has(key uint32) bool {
	fmt.Printf("Has: key %d %b\n", key, key)
	bucket := h.bucket(key)
	fmt.Println("Has: bucket", bucket)
	for _, r := range h.buckets[bucket] {
		if r.key == key {
			return true
		}
	}

	return false
}

func (h *Hash) bucket(hash uint32) uint32 {
	prevBucketCnt := uint32(1 << (h.bits - 1))
	mask := uint32((1 << h.bits) - 1)
	bucket := hash & mask

	// fmt.Println("bucket:", bucket, "<=", h.bucketToSplit, "bits", h.bits, "mask", mask)

	if bucket <= h.bucketToSplit || bucket >= prevBucketCnt {
		return bucket
	}

	return bucket ^ (1 << (h.bits - 1)) // unset the top bit
}

func (h *Hash) split() {
	if h.bucketToSplit == 0 {
		h.bits++
	}

	me := h.bucketToSplit
	mask := uint32((1 << h.bits) - 1)
	fmt.Println("split: me", me, "mask", mask, "bits", h.bits)

	var mine, their []*Record
	for _, r := range h.buckets[me] {
		if r.hash&mask == me {
			mine = append(mine, r)
		} else {
			their = append(their, r)
		}
	}

	h.buckets[me] = mine
	h.buckets = append(h.buckets, their)

	h.bucketToSplit++
	if h.bucketToSplit >= (1 << (h.bits - 1)) {
		h.bucketToSplit = 0
	}

	fmt.Println("split: h.bucketToSplit", h.bucketToSplit, "bits", h.bits)
}

func (h *Hash) Dump() {
	fmt.Println("----hash dump start----")
	fmt.Println("bits", h.bits)
	fmt.Println("records", h.records)
	fmt.Println("bucketToSplit", h.bucketToSplit)

	for i, bucket := range h.buckets {
		var hashes []uint32
		for _, r := range bucket {
			hashes = append(hashes, r.hash)
		}

		fmt.Println("bucket", i, ":", hashes)
	}
	fmt.Println("----hash dump end----\n\n")
}
