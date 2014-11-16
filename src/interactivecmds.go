package main

import (
	"fmt"
	"strings"

	"github.com/lahemi/stack"
)

var (
	fetchTitleState = true
)

// Change this to use the map in cmds.go, somehow.
func doCmd(cmd string, sstack *stack.Stack) {
	switch cmd {
	case "toggleTitleFetch":
		fetchTitleState = !fetchTitleState

	case "println":
		fmt.Println(sstack)

	case "send":
		chanarg, err := sstack.PopE()
		if err != nil {
			stderr(err)
			return
		}
		strarg, err := sstack.PopE()
		if err != nil {
			stderr(err)
			return
		}
		sendToCan(chanarg.(string), strarg.(string))

	case "cmd":
		strarg, err := sstack.PopE()
		if err != nil {
			stderr(err)
			return
		}
		switch arg := strarg.(string); arg {
		case "hello", "emote", "nope":
			fn := CMDS[arg].(func() string)
			sstack.Push(fn())

		case "fortune", "epigram", "callang":
			cargs, err := sstack.PopE()
			if err != nil {
				stderr(err)
				return
			}
			fn := CMDS[arg].(func(string) string)
			sstack.Push(fn(cargs.(string)))
		}
	}
}

func parseEtEval(text string) (ret []string) {
	const (
		RD = iota
		STR
	)
	var (
		spl = strings.Split(text, "") // for UTF-8

		stringmarker = "`"
		sstack       = stack.Stack{}
		buf          string
		state        = RD
	)
	spl = append(spl, " ") // A "terminating" whitespace.

	for i := 0; i < len(spl); i++ {
		c := spl[i]
		switch state {
		case RD:
			switch {
			case isWhite(c) && buf == "":

			case c == stringmarker:
				state = STR
				buf = ""

			case isWhite(c) && buf != "":
				doCmd(buf, &sstack)
				buf = ""

			default:
				buf += c
			}

		case STR:
			switch {
			// So you can use the `stringmarker` in a string.
			case i < len(spl)-1 && c == `\` && spl[i+1] == stringmarker:
				i++
				continue

			case c == stringmarker:
				sstack.Push(buf)
				buf = ""
				state = RD

			default:
				buf += c
			}
		}
	}

	return
}
