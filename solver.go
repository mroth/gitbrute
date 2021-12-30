package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var (
	start     = time.Now()
	startUnix = start.Unix()
)

func Solve(obj []byte) solution {
	possibilities := make(chan try, 512)
	go explore(possibilities)

	winner := make(chan solution)
	done := make(chan struct{})
	for i := 0; i < *cpu; i++ {
		go bruteForce(obj, winner, possibilities, done)
	}
	w := <-winner
	close(done)
	return w
}

type solution struct {
	author, committer date
}

// try is a pair of seconds behind now to brute force, looking for a
// matching commit.
type try struct {
	commitBehind int
	authorBehind int
}

// explore yields the sequence:
//     (0, 0)
//
//     (0, 1)
//     (1, 0)
//     (1, 1)
//
//     (0, 2)
//     (1, 2)
//     (2, 0)
//     (2, 1)
//     (2, 2)
//
//     ...
func explore(c chan<- try) {
	for max := 0; ; max++ {
		for i := 0; i <= max-1; i++ {
			c <- try{i, max}
		}
		for j := 0; j <= max; j++ {
			c <- try{max, j}
		}
	}
}

func bruteForce(obj []byte, winner chan<- solution, possibilities <-chan try, done <-chan struct{}) {
	// blob is the blob to mutate in-place repeatedly while testing
	// whether we have a match.
	blob := []byte(fmt.Sprintf("commit %d\x00%s", len(obj), obj))
	authorDate, adatei := getDate(blob, authorDateRx)
	commitDate, cdatei := getDate(blob, committerDateRx)

	s1 := sha1.New()
	wantHexPrefix := []byte(strings.ToLower(*prefix))
	hexBuf := make([]byte, 0, sha1.Size*2)

	for t := range possibilities {
		select {
		case <-done:
			return
		default:
			ad := date{startUnix - int64(t.authorBehind), authorDate.tz}
			cd := date{startUnix - int64(t.commitBehind), commitDate.tz}
			strconv.AppendInt(blob[:adatei], ad.n, 10)
			strconv.AppendInt(blob[:cdatei], cd.n, 10)
			s1.Reset()
			s1.Write(blob)
			if !bytes.HasPrefix(hexInPlace(s1.Sum(hexBuf[:0])), wantHexPrefix) {
				continue
			}

			winner <- solution{ad, cd}
			return
		}
	}
}

// hexInPlace takes a slice of binary data and returns the same slice with double
// its length, hex-ified in-place.
func hexInPlace(v []byte) []byte {
	const hex = "0123456789abcdef"
	h := v[:len(v)*2]
	for i := len(v) - 1; i >= 0; i-- {
		b := v[i]
		h[i*2+0] = hex[b>>4]
		h[i*2+1] = hex[b&0xf]
	}
	return h
}
