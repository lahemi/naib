package main

import (
	"math/rand"
	"time"
)

type Emoticon struct {
	Cons []string
}

func (e *Emoticon) Pick() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return e.Cons[r.Intn(len(e.Cons))]
}

var (
	DIED = Emoticon{
		Cons: []string{
			"(´Д｀)", "(´∩｀)",
			";_;", "T.T",
		},
	}

	HELLO = Emoticon{
		Cons: []string{
			"ㅇㅅㅇ", "´･ᴗ･`",
			"・ω・", "＾ω＾",
		},
	}

	EMOTES = Emoticon{
		Cons: []string{
			"(;´Д`)", "(｀ω´)", "(｀ε´)",
			"(´･_･`)", "[・ヘ・?]", "[・_・?]",
			"(・◇・)", "(°ヘ°)", "(^～^;)",
			"(´ω｀)", "(×◇×)", "(;°Д°)",
			"L(・o・)」", "(^w^)", "(^_^;)",
			"(^ω^)", "(-＿-)", "(；へ：)",
			"(￣～￣)", "（￣へ￣）", "(￣ω￣;)",
			"(｀＾´)", "(´^｀)", "ヽ(#`Д´)ﾉ",
			"(>_<)", "（＞д＜）", "(¬_¬)",
			"(⇀‸↼‶)", "(°<°)", "~(-△-~)",
			"(~‾▿‾)~", "┌(・。・)┘♪", "⌐■_■",
			"(＾ｖ＾)", "(´～`)", "(^∇^)",
			"( ･᷄ㅂ･᷅ )", "(>_<。)", "(+_+)",
			"(o_-)", "ヘ(°￢°)ノ", "(^◇^)",
			"＜(。_。)＞",
		},
	}

	NOPES = Emoticon{
		Cons: []string{
			"༼ ༎ຶ ෴ ༎ຶ༽",
			"༼ ༏༏ີཻ༾ﾍ ༏༏ີཻ༾༾༽༽",
			"༼´༎ຶ ༎ຶ༽",
			"ヽ༼ ಠ益ಠ ༽ﾉ",
		},
	}
)
