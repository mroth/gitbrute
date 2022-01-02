package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func Solve(obj []byte, prefix string) solution {
	startUnix := time.Now().Unix() // ts to begin looking for matching commits

	// NOTE: parentCtx could be extended to take a ctx in Solve function args to
	// allow for external cancellation, but if doing so we'll also need to modify
	// the exit path and function signature to allow for the possibility of error
	// on a timeout.
	parentCtx := context.Background()
	ctx, cancelWorkers := context.WithCancel(parentCtx)

	winner := make(chan solution)
	explorers := splitExplore(*cpu)
	for _, exf := range explorers {
		exf := exf
		c := newChecker(obj, prefix, startUnix)
		go solver(ctx, exf, c, winner)
	}

	w := <-winner
	cancelWorkers()
	return w
}

func solver(ctx context.Context, exf exploreFunc, c checker, winner chan<- solution) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			t := exf() // next value from explore generator
			if s, ok := c.check(t); ok {
				winner <- s
				return
			}
		}
	}
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
