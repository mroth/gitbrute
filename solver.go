package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Solve(obj []byte, prefix string) solution {
	startUnix := time.Now().Unix() // ts to begin looking for matching commits

	possibilities := make(chan try, 512)
	go explore(possibilities)

	winner := make(chan solution)
	done := make(chan struct{})
	for i := 0; i < *cpu; i++ {
		go bruteForce(obj, prefix, startUnix, winner, possibilities, done)
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

func bruteForce(obj []byte, prefix string, start int64, winner chan<- solution, possibilities <-chan try, done <-chan struct{}) {
	c := newChecker(obj, prefix, start)

	for t := range possibilities {
		select {
		case <-done:
			return
		default:
			if s, ok := c.check(t); ok {
				winner <- s
				return
			}
		}
	}
}

type checker struct {
	wantHexPrefix []byte // desired hex prefix, lowercase
	startUnix     int64  // time to begin search at

	blob                   []byte // storage for mutating obj in place
	authorDate, commitDate date   // original dates extracted from git header
	adatei, cdatei         int    // index of date location in blob
	hexBuf                 []byte // reusable buffer for hex encoding storage
}

func newChecker(obj []byte, prefix string, startUnix int64) checker {
	blob := []byte(fmt.Sprintf("commit %d\x00%s", len(obj), obj))
	authorDate, adatei := getDate(blob, authorDateRx)
	commitDate, cdatei := getDate(blob, committerDateRx)

	return checker{
		wantHexPrefix: []byte(strings.ToLower(prefix)),
		startUnix:     startUnix,
		blob:          blob,
		authorDate:    authorDate,
		commitDate:    commitDate,
		adatei:        adatei,
		cdatei:        cdatei,
		hexBuf:        make([]byte, hex.EncodedLen(sha1.Size)),
	}
}

func (c *checker) check(t try) (newdate solution, ok bool) {
	// mutate blob in place, reusing structures to avoid allocation
	newdate.author = date{c.startUnix - int64(t.authorBehind), c.authorDate.tz}
	newdate.committer = date{c.startUnix - int64(t.commitBehind), c.commitDate.tz}
	strconv.AppendInt(c.blob[:c.adatei], newdate.author.n, 10)
	strconv.AppendInt(c.blob[:c.cdatei], newdate.committer.n, 10)

	sum := sha1.Sum(c.blob)
	hex.Encode(c.hexBuf, sum[:])
	return newdate, bytes.HasPrefix(c.hexBuf, c.wantHexPrefix)
}
