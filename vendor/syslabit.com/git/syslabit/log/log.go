package log

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

type Level string

const (
	// LevelDebug ...
	LevelDebug Level = "DEBUG"

	// LevelInfo ...
	LevelInfo Level = "INFO"

	// LevelWarning ...
	LevelWarning Level = "WARN"

	// LevelError ...
	LevelError Level = "ERROR"

	// LevelFatal ...
	LevelFatal Level = "FATAL"
)

// TraceS ...
type TraceS struct {
	FilePath     string
	LineNr       int
	FunctionName string
}

// RecordS ...
type RecordS struct {
	// ContextID string        `yaml:"context-id"`
	Level     Level         `yaml:"level"`
	Message   string        `yaml:"message"`
	Vars      []interface{} `yaml:"vars"`
	Traceback []TraceS      `yaml:"traceback"`
}

type entryDebugS struct {
	Log       map[string]string
	Traceback []string
}

// Vars ...
type Vars map[string]interface{}

// Debug ...
func Debug(in ...interface{}) *RecordS { return entryCreate(LevelDebug, in) }

// Info ...
func Info(in ...interface{}) *RecordS { return entryCreate(LevelInfo, in) }

// Warning ...
func Warning(in ...interface{}) *RecordS { return entryCreate(LevelWarning, in) }

// Error ...
func Error(in ...interface{}) *RecordS { return entryCreate(LevelError, in) }

// Fatal ...
func Fatal(in ...interface{}) *RecordS {
	if entryCreate(LevelFatal, in) != nil {
		os.Exit(1)
	}
	return nil
}

var framesToIgnore = []string{
	"runtime.goexit",
	"runtime.main",
	"net/http.(*conn).serve",
	"net/http.(*ServeMux).ServeHTTP",
	"net/http.HandlerFunc.ServeHTTP",
	"net/http.serverHandler.ServeHTTP",
}

func entryCreate(level Level, in []interface{}) *RecordS {

	// ukrycie logow na poziomie DEBUG
	if level == LevelDebug && os.Getenv("LOG_SHOW_DEBUG") == "" {
		return nil
	}

	inLen := len(in)
	if inLen == 0 {
		return nil
	}
	if in[0] == nil || in[0] == "" {
		return nil
	}

	// ----------------------------------------------------
	// creating new log entry

	// goID := goroutineID()
	entry := &RecordS{
		// ID:        goID + "-" + common.RandString(8) + "-" + string(level[0]),
		// ContextID: goID,
		Level:     level,
		Message:   fmt.Sprint(in[0]),
		Vars:      in[1:],
		Traceback: Trace(),
	}

	// ----------------------------------------------------
	// print JSON to stdout

	out := map[string]string{
		// "context-id": entry.ContextID,
		"level":   string(entry.Level),
		"message": entry.Message,
	}

	traceback := []string{}
	for _, v := range entry.Traceback {
		traceback = append(traceback, fmt.Sprintf("%s:%d / %s", v.FilePath, v.LineNr, v.FunctionName))
	}

	for i, v := range entry.Vars {
		switch vt := v.(type) {
		case string:
			out[fmt.Sprintf("var/%d", i)] = vt

		case Vars:
			for k2, v2 := range vt {
				if k2 == "context-id" {
					k2 = "var/context-id"
				}
				if k2 == "level" {
					k2 = "var/level"
				}
				if k2 == "message" {
					k2 = "var/message"
				}
				if k2 == "traceback" {
					k2 = "var/traceback"
				}
				out[k2] = fmt.Sprint(v2)
			}

		default:
			out[fmt.Sprintf("var/%d", i)] = fmt.Sprint(v)
		}
	}

	// ----------------------------------------------------
	// message template

	if strings.Contains(entry.Message, "{{") {
		for k, v := range out {
			entry.Message = strings.ReplaceAll(entry.Message, "{{"+k+"}}", v)
		}
		out["message"] = entry.Message
	}

	// ----------------------------------------------------

	switch strings.TrimSpace(strings.ToLower(os.Getenv("LOG_MODE"))) {
	case "simple", "oneline":
		// bardzo prosty jedno-linikowy tryb wyswietlania logów, pozbawiony traceback i varsów
		fmt.Println(time.Now().String()[:23], entry.Level[:1], " ", entry.Message)

	case "multiline-json", "mjson":
		// print multi-line JSON
		buf, _ := json.MarshalIndent(entryDebugS{
			Log:       out,
			Traceback: traceback,
		}, "", "  ")
		fmt.Println(string(buf))

	default:
		// jedno-linikowy json
		out["traceback"] = strings.Join(traceback, "\n")
		buf, _ := json.Marshal(out)
		fmt.Println(string(buf))
	}

	// ----------------------------------------------------

	return entry
}

// --------------------------------------------------

// Trace ...
func Trace() []TraceS {
	trace := []TraceS{}
	pc := make([]uintptr, 10)
	n := runtime.Callers(4, pc)
	frames := runtime.CallersFrames(pc[:n])
	for {
		frame, isMore := frames.Next()
		if !isStringInSlice(frame.Function, framesToIgnore) {
			trace = append(trace, TraceS{
				FilePath:     frame.File,
				LineNr:       frame.Line,
				FunctionName: frame.Function,
			})
		}

		if !isMore {
			break
		}
	}
	return trace
}

// --------------------------------------------------

// var (
// 	goroutineIDLock sync.RWMutex
// 	goroutineIDmap  = map[int]string{}
// )

// func goroutineIDraw() int {
// 	var buf [64]byte
// 	n := runtime.Stack(buf[:], false)
// 	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
// 	id, err := strconv.Atoi(idField)
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
// 	}
// 	return id
// }

// // StartNewContext ...
// func StartNewContext() string {
// 	gid := goroutineIDraw()
// 	goroutineIDLock.Lock()
// 	defer goroutineIDLock.Unlock()
// 	goroutineIDmap[gid] = common.RandString(12)
// 	return goroutineIDmap[gid]
// }

// func goroutineID() string {
// 	gid := goroutineIDraw()
// 	goroutineIDLock.RLock()
// 	res, exist := goroutineIDmap[gid]
// 	goroutineIDLock.RUnlock()
// 	if !exist {
// 		res = StartNewContext()
// 	}
// 	return res
// }

// --------------------------------------------------

func isStringInSlice(s string, slice []string) bool {
	for _, ss := range slice {
		if ss == s {
			return true
		}
	}
	return false
}

// --------------------------------------------------

// PrintJSON ...
func PrintJSON(in interface{}) error {
	buf, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(buf))
	return nil
}
