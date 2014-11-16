package main

import (
	"bufio"
	"database/sql"
	"net"
	"os"
	"regexp"
	"strings"

	_ "github.com/mattn/go-sqlite3"

	"github.com/lahemi/stack"
)

var (
	writechan = make(chan string, 512)
	linerex   = regexp.MustCompile("^:[^ ]+?!([^ ]+? ){3}:.+")
	urlrex    = regexp.MustCompile(
		`^https?://(?:www)?[-~.\w]+(:?(?:/[-+~%/.\w]*)*(?:\??[-+=&;%@.\w]*)*(?:#?[-\.!/\\\w]*)*)?$`,
	)
	titlerex = regexp.MustCompile(`(?i:<title>(.*)</title>)`)

	dataDir = os.Getenv("HOME") + "/.crude"

	startUpConfigFile = dataDir + "/naib.conf"

	fortuneFile = dataDir + "/fortunes.txt"
	epiFile     = dataDir + "/epigrams.txt"
	savedURLs   = dataDir + "/savedURLs.txt"
	dbFile      = dataDir + "/links.db"

	overlord       string
	nick           string
	server         string
	port           string
	cmdPrefix      string
	IRCCmdPrefix   string
	channelsToJoin []string
)

func sendToCan(can, line string) {
	writechan <- "PRIVMSG " + can + " :" + line
}

type MsgLine struct {
	Nick, Cmd, Target, Msg string
}

func (ml MsgLine) isCmd() bool {
	if strings.HasPrefix(ml.Msg, cmdPrefix) {
		return true
	}
	return false
}

type DBFields struct {
	url, title string
	timestamp  int64
	category   string
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

func handleBotCmds(s string) {
	if !linerex.MatchString(s) {
		return
	}
	ml := splitMsgLine(s)

	switch ml.isCmd() {
	case true:
		var (
			linest = ml.Msg[len(cmdPrefix):]
			ind    = strings.Index(linest, " ")
			cmd    string
			cargs  string
		)

		if ind == -1 {
			cmd = linest
		} else {
			cmd = linest[:ind]
			cargs = linest[ind+1:]
		}

		// See cmds.go
		if c, ok := CMDS[cmd]; ok {
			switch cmd {
			case "die":
				c.(func(MsgLine))(ml)

			case "hello", "emote", "nope":
				sendToCan(ml.Target, c.(func() string)())

			case "fortune", "epigram", "callang":
				rv := c.(func(string) string)(cargs)
				if rv != "" {
					sendToCan(ml.Target, rv)
				}

			case "save":
				c.(func(string))(cargs)
			}
		}

	default:
		if !strings.Contains(ml.Msg, "http") {
			return
		}
		for _, w := range strings.Split(ml.Msg, " ") {
			if !urlrex.MatchString(w) {
				continue
			}
			title := fetchTitle(w)
			if title != "" {
				if !fetchTitleState {
					continue
				}
				sendToCan(ml.Target, title)
			}
			saveLinksToDB(DBFields{url: w, title: title})
		}
	}
}

// See `interactivecmds.go`
func handleInteractiveCmds(cmdline string) {
	parseEtEval(cmdline)
}

func init() {
	if fi, err := os.Stat(dataDir); err != nil {
		if err := os.MkdirAll(dataDir, 0777); err != nil {
			die("Unable to create " + dataDir)
		}
		stdout("Initialization, data|config dir " + dataDir + " created.")
	} else if !fi.Mode().IsDir() {
		die("There is a file with the same name as the " + dataDir + " already present.")
	}
	if _, err := os.Stat(savedURLs); err != nil {
		fd, err := os.Create(savedURLs)
		if err != nil {
			die("Failed to create " + savedURLs)
		}
		fd.Close() // Ensures the fd will be freed.
	}

	if fi, err := os.Stat(startUpConfigFile); err == nil && fi.Mode().IsRegular() {
		configs := loadStartUpConfig(startUpConfigFile)
		mandatoryConf := func(s stack.Stack) string {
			v, e := s.PopE()
			if e != nil {
				die("One of the mandatory configuration options not set.")
			}
			return v.(string)
		}
		for k, v := range configs {
			switch k {
			case "nick":
				nick = mandatoryConf(v)

			case "server":
				server = mandatoryConf(v)

			case "port":
				port = mandatoryConf(v)

			case "overlord":
				overlord = mandatoryConf(v)

			case "commandPrefix":
				cmdPrefix = mandatoryConf(v)

			case "IRCCmdPrefix":
				IRCCmdPrefix = mandatoryConf(v)

			case "channels":
				for {
					ch, e := v.PopE()
					if e != nil {
						break
					}
					channelsToJoin = append(channelsToJoin, ch.(string))
				}
			}
		}
	}

	if _, err := os.Stat(dbFile); err != nil {
		db, err := sql.Open("sqlite3", dbFile)
		if err != nil {
			die("Failed to create db file: " + dbFile)
		}
		defer db.Close()
		if _, err = db.Exec(`
            CREATE TABLE IF NOT EXISTS links (
                id INTEGER NOT NULL PRIMARY KEY,
                url TEXT NOT NULL,
                title TEXT,
                timestamp INTEGER NOT NULL, -- UNIX timestamp
                category TEXT,
                FOREIGN KEY(category) REFERENCES categories(category)
            );
            CREATE TABLE IF NOT EXISTS categories (
                category TEXT PRIMARY KEY
            );
            INSERT INTO categories VALUES('music');
            INSERT INTO categories VALUES('img');
            INSERT INTO categories VALUES('lulz');
            INSERT INTO categories VALUES('info');
            INSERT INTO categories VALUES('blank'); -- for no category
        `); err != nil {
			die("Failed to execute SQLite3.")
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", server+":"+port)
	if err != nil {
		die(err)
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
	for _, c := range channelsToJoin {
		writechan <- "JOIN " + c
	}

	// This is so that it's easy to give commands to the
	// bot on the commandline while it's running, no need
	// to do everything through IRC, saving bandwidth.
	for {
		input, err := in.ReadString('\n')
		if err != nil {
			stderr("error input")
			break
		}
		inp := input[:len(input)-1]
		switch {
		case strings.HasPrefix(inp, IRCCmdPrefix):
			writechan <- inp[len(IRCCmdPrefix):]
		default:
			handleInteractiveCmds(inp)
		}
	}
}
