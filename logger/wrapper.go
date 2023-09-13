// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package logger

import (
	"bytes"
	"fmt"
	"os"
	"strings"
)

var (
	Global      Logger
	GlobalLevel = DEBUG
)

var (
	callerSkip = 2
)

func init() {
	Global = NewDefaultLogger(GlobalLevel)
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfigurationFromFile(filename string) error {
	return Global.LoadConfigurationFromFile(filename)
}

// Wrapper for (*Logger).LoadConfiguration
func LoadConfiguration(xc *LogConfig) error {
	return Global.LoadConfiguration(xc)
}

// Wrapper for (*Logger).AddFilter
func AddFilter(name string, lvl level, writer LogWriter) {
	Global.AddFilter(name, lvl, writer)
}

func ChangeFilterLevel(name string, lvl level) {
	Global.ChangeFilterLevel(name, lvl)
}

func GetFilterLevel(name string) level {
	return Global.GetFilterLevel(name)
}

func SetCallerSkip(skip int) {
	callerSkip = skip
}

// Wrapper for (*Logger).Close (closes and removes all logwriters)
func Close() {
	Global.Close()
}

func Crash(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(FATAL, strings.Repeat(" %v", len(args))[1:], args...)
	}
	panic(args)
}

// Logs the given message and crashes the program
func Crashf(format string, args ...interface{}) {
	Global.intLogf(FATAL, format, args...)
	Global.Close() // so that hopefully the messages get logged
	panic(fmt.Sprintf(format, args...))
}

// Compatibility with `log`
func Exit(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

// Compatibility with `log`
func Exitf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
	Global.Close() // so that hopefully the messages get logged
	os.Exit(0)
}

// Compatibility with `log`
func Stderr(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(ERROR, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

// Compatibility with `log`
func Stderrf(format string, args ...interface{}) {
	Global.intLogf(ERROR, format, args...)
}

// Compatibility with `log`
func Stdout(args ...interface{}) {
	if len(args) > 0 {
		Global.intLogf(INFO, strings.Repeat(" %v", len(args))[1:], args...)
	}
}

// Compatibility with `log`
func Stdoutf(format string, args ...interface{}) {
	Global.intLogf(INFO, format, args...)
}

// Send a log message manually
// Wrapper for (*Logger).Log
func Log(lvl level, source, message string) {
	Global.Log(lvl, source, message)
}

// Send a formatted log message easily
// Wrapper for (*Logger).Logf
func Logf(lvl level, format string, args ...interface{}) {
	Global.intLogf(lvl, format, args...)
}

// Send a closure log message
// Wrapper for (*Logger).Logc
func Logc(lvl level, closure func() string) {
	Global.intLogc(lvl, closure)
}

// Utility for finest log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Finest
func Finest(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINEST
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for fine log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Fine
func Fine(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FINE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

// Utility for debug log messages
// When given a string as the first argument, this behaves like Logf but with the DEBUG log level (e.g. the first argument is interpreted as a format for the latter arguments)
// When given a closure of type func()string, this logs the string returned by the closure iff it will be logged.  The closure runs at most one time.
// When given anything else, the log message will be each of the arguments formatted with %v and separated by spaces (ala Sprint).
// Wrapper for (*Logger).Debug
func Debug(arg0 interface{}, args ...interface{}) {
	const (
		lvl = DEBUG
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func DebugW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(DEBUG, b.String())
	}
}

// Utility for trace log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Trace
func Trace(arg0 interface{}, args ...interface{}) {
	const (
		lvl = TRACE
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func TraceW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(TRACE, b.String())
	}
}

// Utility for info log messages (see Debug() for parameter explanation)
// Wrapper for (*Logger).Info
func Info(arg0 interface{}, args ...interface{}) {
	const (
		lvl = INFO
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		Global.intLogc(lvl, first)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(arg0)+strings.Repeat(" %v", len(args)), args...)
	}
}

func InfoW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(INFO, b.String())
	}
}

// Utility for warn log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Warn
func Warn(arg0 interface{}, args ...interface{}) {
	const (
		lvl = WARNING
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func WarnW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(WARNING, b.String())
	}
}

// Utility for error log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Error
func Error(arg0 interface{}, args ...interface{}) {
	const (
		lvl = ERROR
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func ErrorW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(ERROR, b.String())
	}
}

// Utility for critical log messages (returns an error for easy function returns) (see Debug() for parameter explanation)
// These functions will execute a closure exactly once, to build the error message for the return
// Wrapper for (*Logger).Fatal
func Fatal(arg0 interface{}, args ...interface{}) {
	const (
		lvl = FATAL
	)
	switch first := arg0.(type) {
	case string:
		// Use the string as a format string
		Global.intLogf(lvl, first, args...)
	case func() string:
		// Log the closure (no other arguments used)
		str := first()
		Global.intLogf(lvl, "%s", str)
	default:
		// Build a format string so that it will be similar to Sprint
		Global.intLogf(lvl, fmt.Sprint(first)+strings.Repeat(" %v", len(args)), args...)
	}
}

func FatalW(content string, kv ...interface{}) {
	if len(kv)%2 == 0 {
		b := bytes.NewBufferString(content)
		for i := 0; i < len(kv); i += 2 {
			b.WriteString(fmt.Sprintf(" %s=%v", kv[i], kv[i+1]))
		}
		Global.intLogf(FATAL, b.String())
	}
}
