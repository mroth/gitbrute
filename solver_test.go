package main

import (
	"testing"
)

const testObj = `tree a9285ff39f402c54a739037ccae28d81e91bcb56
parent 029691b990bcc859f0f07036e48d2c7f8f2cb329
author Roland Illig <roland@fake.test> 1640342275 +0100
committer Brad Fitzpatrick <brad@fake.test> 1640377083 -0800

Lowercase the prefix from the command line

Previously, running gitbrute with an uppercase hex prefix had resulted
in an endless loop.
`
const testPrefix = "abcd"
const testTS = 1641066267

var testObjBytes = []byte(testObj)

// Realistic approximation of what Solve does for benchmarking purposes,
// breaking out the check step so we can treat it as a RunParallel iteration,
// and continuing regardless of whether a match is found.
func BenchmarkSolveParallel(b *testing.B) {
	b.SetBytes(int64(len(testObjBytes)))

	possibilities := make(chan try, 512)
	go explore(possibilities)

	b.RunParallel(func(pb *testing.PB) {
		c := newChecker(testObjBytes, testPrefix, testTS)
		for pb.Next() {
			t := <-possibilities
			_, _ = c.check(t)
		}
	})
}

// benchmark just the comparison portion, removing any overhead from candidate
// generation and worker coordination.
func BenchmarkCheck(b *testing.B) {
	b.SetBytes(int64(len(testObjBytes)))

	c := newChecker(testObjBytes, testPrefix, testTS)
	t := try{10, 10}
	for i := 0; i < b.N; i++ {
		c.check(t)
	}
}

// benchmark just the comparison portion, removing any overhead from candidate
// generation and worker coordination.
func BenchmarkCheckParallel(b *testing.B) {
	b.SetBytes(int64(len(testObjBytes)))

	b.RunParallel(func(pb *testing.PB) {
		c := newChecker(testObjBytes, testPrefix, testTS)
		t := try{10, 10}
		for pb.Next() {
			_, _ = c.check(t)
		}
	})
}
