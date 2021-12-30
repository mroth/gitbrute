/*
Copyright 2014 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
)

var (
	prefix = flag.String("prefix", "bf", "Desired prefix")
	force  = flag.Bool("force", false, "Re-run, even if current hash matches prefix")
	cpu    = flag.Int("cpus", runtime.NumCPU(), "Number of CPUs to use. Defaults to number of processors.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*cpu)
	if _, err := strconv.ParseInt(*prefix, 16, 64); err != nil {
		log.Fatalf("Prefix %q isn't hex.", *prefix)
	}

	// get hash of current git HEAD
	hash, err := currentHash()
	if err != nil {
		log.Fatal(err)
	}
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
	w := Solve(obj)

	// amend the most recent commit with the skewed timestamps
	cmd := exec.Command("git", "commit", "--allow-empty", "--amend", "--date="+w.author.String(), "--file=-")
	cmd.Env = append(os.Environ(), "GIT_COMMITTER_DATE="+w.committer.String())
	cmd.Stdout = os.Stdout
	cmd.Stdin = bytes.NewReader(msg)
	if err := cmd.Run(); err != nil {
		log.Fatalf("amend: %v", err)
	}
}
