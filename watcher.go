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

	// Basic Colour Constants
	BasicColourRed         = "\033[00;31m"
	BasicColourRedBold     = "\033[01;31m"
	BasicColourGreen       = "\033[00;32m"
	BasicColourGreenBold   = "\033[01;32m"
	BasicColourYellow      = "\033[00;33m"
	BasicColourYellowBold  = "\033[01;33m"
	BasicColourBlue        = "\033[00;34m"
	BasicColourBlueBold    = "\033[01;34m"
	BasicColourMagenta     = "\033[00;35m"
	BasicColourMagentaBold = "\033[01;35m"
	BasicColourCyan        = "\033[00;36m"
	BasicColourCyanBold    = "\033[01;36m"
	BasicColourWhite       = "\033[00;37m"
	BasicColourWhiteBold   = "\033[01;37m"
	ColourReset            = "\033[0m"
)

type Watcher struct {
	interval    time.Duration
	descColour  string
	valueColour string
	logColour   string
	hight       uint16
	width       uint16

	whHardSet bool
}
type cos interface {
	constraints.Ordered
}

var (
	vars  = safe.SortedMap[string, any]{}
	funcs = safe.SortedMap[string, func() any]{}
)

func Create(interval time.Duration) *Watcher {
	os.MkdirAll("/tmp/peek-var", 0755)
	wa := &Watcher{
		interval:    interval,
		descColour:  BasicColourGreenBold,
		valueColour: BasicColourBlueBold,
		logColour:   BasicColourWhiteBold,
	}
	go wa.render()
	return wa
}

func (wa *Watcher) render() {
	var out string
	var wSize *unix.Winsize = &unix.Winsize{
		Row: wa.hight,
		Col: wa.width,
	}
	var buff string
	var sl []string
	var combinedLen int
	var err error
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	go read(r, &buff)

	os.Stdout = w
	for {
		if !wa.whHardSet {
			wSize, err = unix.IoctlGetWinsize(int(oldStdout.Fd()), unix.TIOCGWINSZ)
			if err != nil {
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
		sl = strings.Split(buff, "\n")
		combinedLen = vars.Len() + funcs.Len()
		if len(sl)+combinedLen > int(wSize.Row) {
			buff = strings.Join(sl[len(sl)-int(wSize.Row)+combinedLen:], "\n")
			out += fmt.Sprintf("\033[1;37m%s", buff)
		} else {
			out += fmt.Sprintf("\033[1;37m%s", buff)
		}
		oldStdout.Write([]byte("\033[H\033[2J"))
		oldStdout.Write([]byte(out))
		time.Sleep(wa.interval)
	}
}

func read(r *os.File, buff *string) {
	logFile, _ := os.Create("/tmp/peek-var/log.txt")
	for {
		b := make([]byte, 1024)
		r.Read(b)
		for i := len(b) - 1; i > 0; i-- {
			if b[i] == '\n' {
				*buff += string(b[:i+1])
				logFile.Write(b[:i+1])
				break
			}
		}
	}
}

// Add adds a variable to the watcher.
// description is a string that will be printed before the variable.
// Variable can be any type that implements the 'cos' interface.
func Var[T cos](desc string, v *T) {
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
		wa.descColour = BasicColourGreenBold
	}
	if wa.valueColour == "" {
		wa.valueColour = BasicColourBlueBold
	}
	if wa.logColour == "" {
		wa.logColour = BasicColourWhiteBold
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
