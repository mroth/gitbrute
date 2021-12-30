package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// currentHash retrieves the hash of current HEAD via git.
func currentHash() (string, error) {
	all, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "", err
	}
	h := string(all)
	if i := strings.Index(h, "\n"); i > 0 {
		h = h[:i]
	}
	return h, nil
}

// extractMessage returns only the message portion of a commit object.
func extractCommitMessage(obj []byte) ([]byte, error) {
	i := bytes.Index(obj, []byte("\n\n"))
	if i < 0 {
		return nil, fmt.Errorf("no \\n\\n found in %q", obj)
	}
	msg := obj[i+2:]
	return msg, nil
}

// date is a git date.
type date struct {
	n  int64 // unix seconds
	tz string
}

func (d date) String() string { return fmt.Sprintf("%d %s", d.n, d.tz) }

var (
	authorDateRx    = regexp.MustCompile(`(?m)^author.+> (.+)`)
	committerDateRx = regexp.MustCompile(`(?m)^committer.+> (.+)`)
)

// getDate parses out a date from a git header (or blob with a header
// following the size and null byte). It returns the date and index
// that the unix seconds begins at within h.
func getDate(h []byte, rx *regexp.Regexp) (d date, idx int) {
	m := rx.FindSubmatchIndex(h)
	if m == nil {
		log.Fatalf("Failed to match %s in %q", rx, h)
	}
	v := string(h[m[2]:m[3]])
	space := strings.Index(v, " ")
	if space < 0 {
		log.Fatalf("unexpected date %q", v)
	}
	n, err := strconv.ParseInt(v[:space], 10, 64)
	if err != nil {
		log.Fatalf("unexpected date %q", v)
	}
	return date{n, v[space+1:]}, m[2]
}
