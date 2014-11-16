package main

import (
	"os"
	"os/exec"
	"time"
)

var CMDS = map[string]func(MsgLine, string){
	"die": func(ml MsgLine, cargs string) {
		if ml.Nick == overlord {
			sendToCan(ml.Target, DIED.Pick())
			time.Sleep(time.Duration(1) * time.Second)
			writechan <- "QUIT"
			os.Exit(0)
		}
	},

	"hello": func(ml MsgLine, cargs string) {
		sendToCan(ml.Target, HELLO.Pick())
	},

	"emote": func(ml MsgLine, cargs string) {
		sendToCan(ml.Target, EMOTES.Pick())
	},

	"nope": func(ml MsgLine, cargs string) {
		sendToCan(ml.Target, NOPES.Pick())
	},

	"fortune": func(ml MsgLine, cargs string) {
		if fort := lineWithOptionalMatch(fortuneFile, cargs); fort != "" {
			sendToCan(ml.Target, fort)
		}
	},

	"epigram": func(ml MsgLine, cargs string) {
		if epigram := lineWithOptionalMatch(epiFile, cargs); epigram != "" {
			sendToCan(ml.Target, epigram)
		}
	},

	// Silly and incomplete, feel free to remove.
	// Uses https://github.com/lahemi/gochunks/tree/master/callang
	"callang": func(ml MsgLine, cargs string) {
		out, err := exec.Command("callang", "-s", cargs).Output()
		if err != nil || string(out) == "" {
			return
		}
		sendToCan(ml.Target, string(out))
	},

	"save": func(ml MsgLine, cargs string) {
		if out := saveURL(cargs); out != "" {
			stdout(out)
		}
	},
}
