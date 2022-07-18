package example

import (
	"fmt"
	"strings"
	"time"

	peak "github.com/fr-str/var-peek"
)

func main() {
	idx := 0
	peak.Create(100 * time.Millisecond)
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
