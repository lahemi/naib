package main

import (
	"fmt"
	"strings"

	"github.com/lahemi/stack"
)

var (
	fetchTitleState = true
)

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

func parse(text string) (ret []string) {
	var (
		spl = strings.Split(text, "") // for utf-8 chars

		sexpS = "("
		sexpE = ")"
		buf   string
	)

	for i := 0; i < len(spl); i++ {
		c := spl[i]
		switch {
		case isWhite(c) && buf != "":
			ret = append(ret, buf)
			buf = ""
		case isWhite(c):
		case c == sexpS || c == sexpE:
			if buf != "" {
				ret = append(ret, buf)
				buf = ""
			}
			ret = append(ret, c)
		default:
			buf += c
		}
	}

	return
}

// Need to be rewritten due to silly, unclear and fragile code.
func eval(ss []string) {
	var (
		sexpS     = "("
		sexpE     = ")"
		last_expr string

		cmd    string
		args   []string
		istack = stack.Stack{}
	)

	for i := 0; i < len(ss); i++ {
		s := ss[i]
		switch {
		case s == sexpS:
			last_expr = sexpS
			istack.Push(i)
		case s == sexpE && last_expr == sexpS:
			m, e := istack.PopE()
			if e != nil {
				break
			}
			n := m.(int)
			cmd = ss[n+1]
			args = ss[n+2 : i]
			r := doCmd(cmd, args)
			ss = ss[:n+1]
			if r != "" {
				ss[n] = r
			}
			i = 0
			ss = append(ss, ")")
		}
	}
}
