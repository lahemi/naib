package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func clock() string {
	h, m, s := time.Now().Clock()
	return fmt.Sprintf("[%d %d %d]", h, m, s)
}

func stdout(str ...interface{}) {
	fmt.Fprintf(os.Stdout, "%s %v\n", clock(), str)
}

func stderr(str ...interface{}) {
	fmt.Fprintf(os.Stderr, "%s %v\n", clock(), str)
}

func die(str ...interface{}) {
	stderr(str)
	os.Exit(1)
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

// "What was that one site..."
// Let's save some urls and titles
// - Antti-Ville Jokela
func saveUrl(url, file string) string {
	// First, let's get that title
	if len(url) < 2 {
		return "No url found"
	}

	// One option is to check here that title is
	// not empty - will be ignored for now
	url = strings.Trim(url, " ")
	title := fetchTitle(url)

	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		stderr(err)
		return "Failed to open file: " + file
	}
	defer f.Close()

	content := title + " : " + url + "\n"

	if _, err = f.WriteString(content); err != nil {
		stderr(err)
		return "Write failed."
	}

	return "Url '" + title + "' is now saved"
}

func fetchTitle(msgWord string) string {
	resp, err := http.Get(msgWord)
	if err != nil {
		stderr("Nope at GETing " + msgWord)
		return ""
	}
	val := resp.Header.Get("Content-Type")
	if val == "" || !strings.Contains(val, "text/html") {
		return ""
	}
	var buf string
	reader := bufio.NewReader(resp.Body)
	for {
		word, err := reader.ReadBytes(' ')
		if err != nil {
			stderr("Nope at reading the site " + string(word))
			return ""
		}
		if err == io.EOF {
			break
		}
		buf += string(word)
		if m, _ := regexp.MatchString(".*(?i:</title>).*?", string(word)); m {
			break
		}
		if len(buf) > 8192 {
			break
		}
	}
	titleMatch := titlerex.FindStringSubmatch(buf)
	if len(titleMatch) == 2 {
		stdout(len(buf))
		return titleMatch[1]
	} else {
		stdout("No title found")
		return ""
	}
}

func saveLinksToDB(fields DBFields) {
	// Pretty funky, pattern matching-like.
	switch "" {
	case fields.url:
		stderr("No url to save to DB.")
		return
	case fields.title:
		fields.title = "blank"
	case fields.category:
		fields.category = "blank"
	}

	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		stderr("Failed to open the DB.")
		return
	}
	defer db.Close()

	fields.timestamp = int64(time.Now().Unix())

	stmt, err := db.Prepare(`
        INSERT INTO links(url, title, timestamp, category)
        VALUES(?, ?, ?, ?)
    `)
	if err != nil {
		stderr("Failed to prepare SQL statement with error:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.Exec(fields.url, fields.title, fields.timestamp, fields.category)
	if err != nil {
		stderr("Failed to write to DB.")
		return
	}
	stdout(fields.url + " saved to DB.")
}
