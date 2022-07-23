package peek

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/Main-Kube/util/safe"
	"golang.org/x/exp/constraints"
	"golang.org/x/sys/unix"
)

const (
	// Font

	//  Colour Constants
	Red         = "\033[00;31m"
	RedBold     = "\033[01;31m"
	Green       = "\033[00;32m"
	GreenBold   = "\033[01;32m"
	Yellow      = "\033[00;33m"
	YellowBold  = "\033[01;33m"
	Blue        = "\033[00;34m"
	BlueBold    = "\033[01;34m"
	Magenta     = "\033[00;35m"
	MagentaBold = "\033[01;35m"
	Cyan        = "\033[00;36m"
	CyanBold    = "\033[01;36m"
	White       = "\033[00;37m"
	WhiteBold   = "\033[01;37m"
	Reset       = "\033[0m"
)

type Watcher struct {
	interval    time.Duration
	descColour  string
	valueColour string
	logColour   string
	hight       uint16
	width       uint16
	buff        string
	b           []byte
	c           chan struct{}

	whHardSet bool
}

var (
	vars       = safe.SortedMap[string, any]{}
	funcs      = safe.SortedMap[string, func() any]{}
	logFile, _ = os.Create("/tmp/peek-var/log.txt")
)

func Create(interval time.Duration) *Watcher {
	os.MkdirAll("/tmp/peek-var", 0755)
	wa := &Watcher{
		interval:    interval,
		descColour:  GreenBold,
		valueColour: BlueBold,
		logColour:   WhiteBold,
		b:           make([]byte, 1024),
		c:           make(chan struct{}, 1),
	}
	go wa.render()
	return wa
}

func (wa *Watcher) render() {
	defer logFile.Close()
	var out string
	var wSize *unix.Winsize = &unix.Winsize{
		Row: wa.hight,
		Col: wa.width,
	}
	var sl []string
	var combinedLen int
	var err error
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	go wa.read(r)

	os.Stdout = w
	for {
		if !wa.whHardSet {
			wSize, err = unix.IoctlGetWinsize(int(oldStdout.Fd()), unix.TIOCGWINSZ)
			if err != nil {
				fmt.Println("Error getting window size:", err)
				fmt.Println("If you are using a non-standard terminal, you can set the window size by running 'wa.SetDimentions(hight, width)'")
				panic(err)
			}
		}

		out = ""
		for v := range vars.Iter() {
			out += fmt.Sprintf("%s%s%s%v\n", wa.descColour, v.Key, wa.valueColour, reflect.Indirect(reflect.ValueOf(v.Value)))
		}

		for v := range funcs.Iter() {
			out += fmt.Sprintf("%s%s%s%v\n", wa.descColour, v.Key, wa.valueColour, v.Value())
		}

		sl = strings.Split(wa.buff, "\n")
		combinedLen = vars.Len() + funcs.Len()
		if len(sl)+combinedLen > int(wSize.Row) {
			wa.buff = strings.Join(sl[len(sl)-int(wSize.Row)+combinedLen:], "\n")
			out += fmt.Sprintf("%s%s", wa.logColour, wa.buff)
		} else {
			out += fmt.Sprintf("%s%s", wa.logColour, wa.buff)
		}
		oldStdout.Write([]byte("\033[H\033[2J"))
		oldStdout.Write([]byte(out))
		// rerenders the screen after the interval or after new data is received
		// I don't know if this is the best idea ¯\_(ツ)_/¯
		select {
		case <-wa.c:
		case <-time.After(wa.interval):
		}
	}
}

func (wa *Watcher) read(r *os.File) {
	for {
		i, _ := r.Read(wa.b)
		wa.buff += string(wa.b[:i])
		logFile.Write(wa.b[:i])
		wa.c <- struct{}{}
	}
}

// Add adds a variable to the watcher.
// description is a string that will be printed before the variable.
// Variable can be any type in 'constraints.ordered' interface.
func Var[T constraints.Ordered](desc string, v *T) {
	vars.Set(desc, v)
}

// Add adds a func to the watcher which will be run each on itteration.
// description is a string that will be printed before the variable returned by func.
func Func(desc string, v func() any) {
	funcs.Set(desc, v)
}

// SetColour sets the colour of the description, value and logs.
func (wa *Watcher) SetColour(desc, value string, log string) {
	wa.descColour = desc
	wa.valueColour = value
	wa.logColour = log
	if wa.descColour == "" {
		wa.descColour = GreenBold
	}
	if wa.valueColour == "" {
		wa.valueColour = BlueBold
	}
	if wa.logColour == "" {
		wa.logColour = WhiteBold
	}

}

// Set one of 256 colours for description, value and logs.
// See https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
func (wa *Watcher) Set256ColourMode(desc, value, log int) {
	wa.descColour = fmt.Sprintf("\033[38;5;%dm", desc)
	wa.valueColour = fmt.Sprintf("\033[38;5;%dm", value)
	wa.logColour = fmt.Sprintf("\033[38;5;%dm", log)
}

// If unix.IoctlGetWinsize is giving you trouble, you can use this function to set the width and height of the window.
// To get the size of the window run 'stty size'
func (wa *Watcher) SetDimentions(h, w int) {
	wa.hight = uint16(h)
	wa.width = uint16(w)
	wa.whHardSet = true
}
