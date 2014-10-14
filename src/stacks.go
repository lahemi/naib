package main

type StrStack []string

func (s *StrStack) Push(str string) {
	(*s) = append((*s), str)
}
func (s *StrStack) Pop() string {
	if len((*s)) <= 0 {
		return ""
	}
	last := (*s)[len((*s))-1]
	(*s) = (*s)[:len((*s))-1]
	return last
}

type IntStack []int

func (s *IntStack) Push(i int) {
	(*s) = append((*s), i)
}
func (s *IntStack) Pop() int {
	if len((*s)) <= 0 {
		return -0xffffffff
	}
	last := (*s)[len((*s))-1]
	(*s) = (*s)[:len((*s))-1]
	return last
}
