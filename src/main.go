package main

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	writechan = make(chan string, 512)
	linerex   = regexp.MustCompile("^:[^ ]+?!([^ ]+? ){3}:.+")
	urlrex    = regexp.MustCompile(
		`^https?://(?:www)?[-~.\w]+(:?(?:/[-+~%/.\w]*)*(?:\??[-+=&;%@.\w]*)*(?:#?[-\.!/\\\w]*)*)?$`,
	)
	titlerex = regexp.MustCompile(`(?i:<title>(.*)</title>)`)

	cmdPrefix         = "˙"
	interactCmdPrefix = "("

	overLord = "" // You

	dataDir = os.Getenv("HOME") + "/.crude"

	fortuneFile = dataDir + "/fortunes.txt"
	epiFile     = dataDir + "/epigrams.txt"
)

func sendToCan(can, line string) {
	writechan <- "PRIVMSG " + can + " :" + line
}

type MsgLine struct {
	Nick, Cmd, Target, Msg string
}

func splitMsgLine(l string) MsgLine {
	spl := strings.SplitN(l, " ", 4)
	return MsgLine{
		Nick:   spl[0][1:strings.Index(l, "!")],
		Cmd:    spl[1],
		Target: spl[2],
		Msg:    spl[3][1:],
	}
}

func handleOut(s string) {
	if linerex.MatchString(s) {
		ml := splitMsgLine(s)
		sep := " | "
		stdout(ml.Nick + sep + ml.Target + sep + ml.Msg)
	} else {
		stdout(s)
	}
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

func handleBotCmds(s string) {
	if !linerex.MatchString(s) {
		return
	}
	ml := splitMsgLine(s)

	if ml.Nick == overLord && ml.Msg == cmdPrefix+"die" {
		sendToCan(ml.Target, DIED.Pick())
		// BAD, use a sync.WaitGroup maybe ?
		time.Sleep(time.Duration(1) * time.Second)
		writechan <- "QUIT"
		os.Exit(0)
	}

	switch {
	case strings.HasPrefix(ml.Msg, cmdPrefix):
		linest := ml.Msg[len(cmdPrefix):]
		switch {
		case linest == "hello":
			sendToCan(ml.Target, HELLO.Pick())
		case linest == "emote":
			sendToCan(ml.Target, EMOTES.Pick())
		case linest == "nope":
			sendToCan(ml.Target, NOPES.Pick())
			// Generalize these two
		case linest == "fortune":
			fort := getFortune(fortuneFile)
			if fort != "" {
				sendToCan(ml.Target, fort)
			}
		case strings.HasPrefix(linest, "epigram"):
			epigram := getEpigram(epiFile, linest[7:])
			if epigram != "" {
				sendToCan(ml.Target, epigram)
			}
        case strings.HasPrefix(linest, "callang"):
            out := doCallang(linest[7:])
            if out != "" {
                sendToCan(ml.Target, out)
            }
		}
	default:
		if !fetchTitleState {
			return
		}
		if !strings.Contains(ml.Msg, "http") {
			return
		}
		for _, w := range strings.Split(ml.Msg, " ") {
			if !urlrex.MatchString(w) {
				continue
			}
			if title := fetchTitle(w); title != "" {
				sendToCan(ml.Target, title)
			}
		}
	}
}

// See `interactivecmds.go`
func handleInteractiveCmds(cmdline string) {
	eval(parse(cmdline))
}

func main() {

	if !checkDataDir(dataDir) {
		if e := os.Mkdir(dataDir, 0755); e != nil {
			stderr("unable to create data dir")
		}
		stdout("Initialization, data|config dir " + dataDir + " created.")
	}

	var (
		server = "irc.freenode.net"
		port   = "6667"
		nick   = "gecannels"
	)

	conn, err := net.Dial("tcp", server+":"+port)
	if err != nil {
		panic(err)
	}
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)
	in := bufio.NewReader(os.Stdin)

	go func() {
		for {
			str, err := r.ReadString('\n')
			if err != nil {
				stderr("read error")
				break
			}
			if str[:4] == "PING" {
				writechan <- "PONG" + str[4:len(str)-2]
			} else {
				handleOut(str[:len(str)-2])
				handleBotCmds(str[:len(str)-2])
			}
		}
	}()
	go func() {
		for {
			str := <-writechan
			if _, err := w.WriteString(str + "\r\n"); err != nil {
				stderr("write error")
				break
			}
			stdout(str)
			w.Flush()
		}
	}()

	writechan <- "USER " + nick + " * * :" + nick
	writechan <- "NICK " + nick

	for {
		input, err := in.ReadString('\n')
		if err != nil {
			stderr("error input")
			break
		}
		inp := input[:len(input)-1]
		switch {
		case strings.HasPrefix(inp, interactCmdPrefix):
			handleInteractiveCmds(inp)
		default:
			writechan <- inp
		}
	}
}
