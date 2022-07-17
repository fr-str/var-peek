package main

import (
	"fmt"
	"peak/peak"
	"strings"
	"time"
)

func main() {
	idx := 0
	w := peak.Create(100 * time.Millisecond)
	w.Run()

	peak.Var("idx: ", &idx)
	peak.Var("xD: ", &idx)
	peak.Var("dupa: ", &idx)
	peak.Var("hehe he: ", &idx)
	peak.Func("time: ", func() any {
		return strings.Split(time.Now().String(), "+")[2]
	})
	for {
		idx++
		fmt.Println("-------------------------", idx, "-------------------------")
		time.Sleep(100 * time.Millisecond)
	}

}
