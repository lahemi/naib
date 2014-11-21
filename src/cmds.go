package main

import (
	"os"
	"os/exec"
	"strings"
	"time"
)

// interface{} makes using these a bit more verbose,
// what with all the type assertions, but on the other
// hand, now it's easier to use same stuff for commands
// coming from both IRC and interactive cmdline mode.
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
	"callang": func(cargs string) string {
		out, err := exec.Command("callang", "-s", cargs).Output()
		if err != nil {
			stderr(err)
			return ""
		}
		return string(out)
	},

	"save": func(cargs string) {
		if out := saveURL(cargs); out != "" {
			stdout(out)
		}
	},

	// Uses github.com/lahemi/portfinder
	// Add some kind of search or alias thing for these ?
	"services": func(cargs string) string {
		if cargs == "" {
			cargs = "pastecan filecan datacan fserve"
		}
		for _, s := range strings.Split(cargs, " ") {
			switch s {
			case "pastecan", "filecan", "datacan", "fserve":
			default:
				stderr("Illegal argument to `services`.")
				return ""
			}
		}
		// silence and tolerate erroneous output, do not show the PID
		cargs = "-s -t -n " + cargs
		out, err := exec.Command("portfinder", strings.Split(cargs, " ")...).Output()
		if err != nil {
			stderr("portfinder failed with:", err)
			return ""
		}
		sout := string(out)
		replaceAmount := strings.Count(sout, "\n") - 1
		return strings.Replace(sout, "\n", " | ", replaceAmount)
	},
}
