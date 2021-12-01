package raft

import (
	"fmt"
	"testing"
)

func TestSimpleLL(t *testing.T) {
	simpleLL := SimpleLL{}
	simpleLL.Add("siema")
	simpleLL.Add("heja")
	simpleLL.Add("co tam")

	fmt.Println(simpleLL.String())
	fmt.Println(simpleLL.GetLast())
}
