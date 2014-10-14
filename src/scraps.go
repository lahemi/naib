package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
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

func getFortune(file string) string {
	cnt, err := ioutil.ReadFile(file)
	if err != nil {
		stderr(err)
		return ""
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	spl := strings.Split(string(cnt), "\n")
	return spl[r.Intn(len(spl)-1)]
}
