package main

import (
	"os"
	"os/exec"
	"time"
)

var CMDS = map[string]interface{}{
	"die": func(ml MsgLine) {
		if ml.Nick == overlord {
			sendToCan(ml.Target, DIED.Pick())
			time.Sleep(time.Duration(1) * time.Second)
			writechan <- "QUIT"
			os.Exit(0)
		}
	},

	"hello": func() string {
		return HELLO.Pick()
	},

	"emote": func() string {
		return EMOTES.Pick()
	},

	"nope": func() string {
		return NOPES.Pick()
	},

	"fortune": func(cargs string) string {
		return lineWithOptionalMatch(fortuneFile, cargs)
	},

	"epigram": func(cargs string) string {
		return lineWithOptionalMatch(epiFile, cargs)
	},

	// Silly and incomplete, feel free to remove.
	// Uses https://github.com/lahemi/gochunks/tree/master/callang
	"callang": func(ml MsgLine, cargs string) string {
		out, err := exec.Command("callang", "-s", cargs).Output()
		if err != nil {
			return ""
		}
		return string(out)
	},

	"save": func(cargs string) {
		if out := saveURL(cargs); out != "" {
			stdout(out)
		}
	},
}
