package main

import (
	"io/ioutil"

	"github.com/lahemi/stack"
)

func loadStartUpConfig(file string) map[string]stack.Stack {
	cnt, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	// States
	const (
		RD = iota
		COMP
		CMT
	)
	var (
		buf    string
		gstack = stack.Stack{}
		text   = []rune(string(cnt))
		state  = RD

		ops = map[string]rune{
			"comp": ':',
			"cmt":  '\\',
		}

		dict = map[string]stack.Stack{}
	)
	bufPush := func() {
		if buf != "" {
			gstack.Push(buf)
			buf = ""
		}
	}
	for i := 0; i < len(text); i++ {
		c := text[i]
		switch state {
		case RD:
			switch c {
			case ops["cmt"]:
				state = CMT
				bufPush()
			case ops["comp"]:
				if buf == "" || (len(buf) > 0 && string(buf[len(buf)-1]) == `\`) {
					state = COMP
					bufPush()
				}
			case ' ', '\t', '\n':
				bufPush()

			default:
				buf += string(c)
			}

		case COMP:
			switch c {
			case '\n':
				if buf == "" {
					die("Invalid config file.")
				}
				dict[buf] = gstack
				gstack = stack.Stack{}
				buf = ""
				state = RD

			case ' ', '\t': // ignore

			default:
				buf += string(c)
			}

		case CMT:
			if c == '\n' {
				state = RD
			}
		}
	}

	return dict
}
