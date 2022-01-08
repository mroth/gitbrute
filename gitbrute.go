// The gitbrute command brute-forces a git commit hash prefix.
package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	prefix  = flag.String("prefix", defaultPrefix(), "Desired prefix. Set default via $GITBRUTE_PREFIX.")
	force   = flag.Bool("force", false, "Re-run, even if current hash already matches prefix.")
	cpu     = flag.Int("cpus", runtime.NumCPU(), "Number of workers to use. Defaults to number of processors.")
	dryrun  = flag.Bool("dryrun", false, "Do not amend commit.")
	verbose = flag.Bool("v", false, "Verbose output.")
)

func defaultPrefix() string {
	if prefix := os.Getenv("GITBRUTE_PREFIX"); prefix != "" {
		return prefix
	}
	return "bf"
}

func main() {
	flag.Parse()
	if _, err := strconv.ParseInt(*prefix, 16, 64); err != nil {
		log.Fatalf("Prefix %q isn't hex.", *prefix)
	}

	// get hash of current git HEAD
	hash, err := currentHash()
	if err != nil {
		log.Fatal(err)
	}
	logVerboseF("Current HEAD hash %v", hash)
	if strings.HasPrefix(hash, *prefix) && !*force {
		return
	}

	// obtain the commit object
	obj, err := exec.Command("git", "cat-file", "-p", hash).Output()
	if err != nil {
		log.Fatal(err)
	}

	// extract commit message from the commit object
	msg, err := extractCommitMessage(obj)
	if err != nil {
		log.Fatal(err)
	}

	// search (forever) until a solution is found
	logVerboseF("Beginning search for hash matching prefix `%v`", *prefix)
	start := time.Now()
	w := Solve(obj, *prefix)
	elapsed := time.Since(start)

	// if -dryrun, show information and exit instead
	logVerboseF("Found matching commit in %v", elapsed)
	if *dryrun {
		logVerboseF("Dry run: exiting without amending...")
		os.Exit(0)
	}

	// amend the most recent commit with the skewed timestamps
	cmd := exec.Command("git", "commit", "--allow-empty", "--amend", "--date="+w.author.String(), "--file=-")
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+w.committer.String())
	cmd.Stdout = os.Stdout
	cmd.Stdin = bytes.NewReader(msg)
	if err := cmd.Run(); err != nil {
		log.Fatalf("amend: %v", err)
	}
}

func logVerboseF(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
}
