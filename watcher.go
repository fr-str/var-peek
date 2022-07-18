package peek

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

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
}
type cos interface {
	constraints.Ordered
}

var (
	vars  = SortedMap[string, any]{}
	funcs = SortedMap[string, func() any]{}
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
	var wSize *unix.Winsize
	var buff string
	var sl []string
	var combinedLen int
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	go read(r, &buff)

	os.Stdout = w
	for {
		wSize, _ = unix.IoctlGetWinsize(int(oldStdout.Fd()), unix.TIOCGWINSZ)

		out = ""
		for v := range vars.Iter() {
			out += fmt.Sprintf("%[3]s%[1]s%[4]s%[2]v", v.Key, reflect.Indirect(reflect.ValueOf(v.Value)), wa.descColour, wa.valueColour) + "\n"
		}
		for v := range funcs.Iter() {
			out += fmt.Sprintf("%[3]s%[1]s%[4]s%[2]v", v.Key, v.Value(), wa.descColour, wa.valueColour) + "\n"
		}
		sl = strings.Split(buff, "\n")
		combinedLen = vars.Len() + funcs.Len()
		if len(sl)+combinedLen > int(wSize.Row) {
			buff = strings.Join(sl[len(sl)-int(wSize.Row)+combinedLen:], "\n")
			out += "\033[1;37m" + buff
		} else {
			out += "\033[1;37m" + buff
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
