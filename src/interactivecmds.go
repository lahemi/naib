package main

import (
	"fmt"
	"strings"
)

var (
	fetchTitleState = true
)

func isWhite(c string) bool {
	if c == " " || c == "\t" || c == "\n" {
		return true
	}
	return false
}

func doCmd(cmd string, args []string) string {
	switch cmd {
	case "toggleTitleFetch":
		fetchTitleState = !fetchTitleState
	case "println":
		fmt.Println(args)
	case "send":
		sendToCan(args[0], strings.Join(args[1:], " "))
	case "cmd":
		switch {
		case args[0] == "hello":
			return HELLO.Pick()
		case args[0] == "emote":
			return EMOTES.Pick()
		case args[0] == "nope":
			return NOPES.Pick()
		}
	}
	return ""
}

func parse(text string) StrStack {
	var (
		spl = strings.Split(text, "") // for utf-8 chars

		sexpS = "("
		sexpE = ")"
		buf   string

		sstack = StrStack{}
	)

	for i := 0; i < len(spl); i++ {
		c := spl[i]
		switch {
		case isWhite(c) && buf != "":
			sstack.Push(buf)
			buf = ""
		case isWhite(c):
		case c == sexpS || c == sexpE:
			if buf != "" {
				sstack.Push(buf)
				buf = ""
			}
			sstack.Push(c)
		default:
			buf += c
		}
	}

	return sstack
}

func eval(sstack StrStack) {
	var (
		sexpS     = "("
		sexpE     = ")"
		last_expr string

		cmd    string
		args   []string
		istack = IntStack{}
	)

	for i := 0; i < len(sstack); i++ {
		switch {
		case sstack[i] == sexpS:
			last_expr = sexpS
			istack.Push(i)
		case sstack[i] == sexpE && last_expr == sexpS:
			m := istack.Pop()
			// Bad error handling here.
			if m == -0xffffffff {
				break
			}
			cmd = sstack[m+1]
			args = sstack[m+2 : i]
			r := doCmd(cmd, args)
			sstack = sstack[:m+1]
			if r != "" {
				sstack[m] = r
			}
			i = 0
			sstack.Push(")")
		}
	}
}
