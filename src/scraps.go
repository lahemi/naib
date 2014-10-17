package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"
)

func clock() string {
	h, m, s := time.Now().Clock()
	return fmt.Sprintf("[%d %d %d]", h, m, s)
}

func stdout(str ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s %v\n", clock(), str)
}

func stderr(str ...interface{}) {
	fmt.Fprintln(os.Stderr, "%s %v\n", clock(), str)
}

func checkDataDir(ddir string) bool {
	if f, e := os.Stat(ddir); e != nil || !f.IsDir() {
		return false
	}
	return true
}

func isWhite(c string) bool {
	if c == " " || c == "\t" || c == "\n" {
		return true
	}
	return false
}

func lineWithOptionalMatch(file, arg string) string {
	cnt, err := ioutil.ReadFile(file)
	if err != nil {
		stderr(err)
		return ""
	}

	var (
		args  = strings.Split(strings.TrimPrefix(arg, " "), " ")
		lines = strings.Split(string(cnt), "\n")
		r     = rand.New(rand.NewSource(time.Now().UnixNano()))
	)

	if len(args) == 0 {
		return lines[r.Intn(len(lines)-1)]
	}

	var (
		matches    []string
		matchCount = 0
		maxMatch   = 0
	)

	for _, ep := range lines {
		matchCount = 0
		for _, a := range args {
			if len(a) > 0 && !isWhite(a) {
				if strings.Contains(strings.ToLower(ep), a) {
					matchCount++
				}
			}
		}
		if matchCount == maxMatch {
			matches = append(matches, ep)
		}
		if matchCount > maxMatch {
			maxMatch = matchCount
			matches = []string{}
			matches = append(matches, ep)
		}
	}
	if len(matches) == 0 {
		return lines[r.Intn(len(lines)-1)]
	}
	if len(matches) == 1 {
		return matches[0]
	}

	return matches[r.Intn(len(matches)-1)]
}

func doCallang(cmd string) string {
	out, err := exec.Command("callang", "-s", cmd).Output()
	if err != nil {
		return ""
	}
	return string(out)
}
