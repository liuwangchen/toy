// Copyright (C) 2010, Kyle Lemons <kyle@kylelemons.net>.  All rights reserved.

package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/liuwangchen/toy/pkg/console"
)

const (
	colorSymbol = 0x1B
)

/*
前景色            背景色           颜色
---------------------------------------
30                40              黑色
31                41              红色
32                42              绿色
33                43              黃色
34                44              蓝色
35                45              紫红色
36                46              青蓝色
37                47              白色
*/
var (
	levelColor = [...]int{30, 30, 37, 37, 32, 34, 31, 35}
)

var stdout io.Writer = os.Stdout

type ConsoleOp struct {
	color bool
}

func (op *ConsoleOp) applyOpts(opts []OpConsoleOp) {
	for _, opt := range opts {
		opt(op)
	}
}

type OpConsoleOp func(*ConsoleOp)

func WithConsoleColor() OpConsoleOp {
	return func(op *ConsoleOp) { op.color = true }
}

// This is the standard writer that prints to standard output.
type ConsoleLogWriter chan *LogRecord

// This creates a new ConsoleLogWriter
func NewConsoleLogWriter(opts ...OpConsoleOp) ConsoleLogWriter {
	op := &ConsoleOp{}
	op.applyOpts(opts)
	records := make(ConsoleLogWriter, LogBufferLength)
	if op.color {
		go records.runWithColor(stdout)
	} else {
		go records.run(stdout)
	}
	return records
}

func NewColorConsoleLogWriter() ConsoleLogWriter {
	return NewConsoleLogWriter(WithConsoleColor())
}

func (w ConsoleLogWriter) run(out io.Writer) {
	var timestr string
	var timestrAt int64

	for rec := range w {
		if at := rec.Created.UnixNano() / 1e9; at != timestrAt {
			timestr, timestrAt = rec.Created.Format("01/02/06 15:04:05"), at
		}
		fmt.Fprint(out, "[", timestr, "] [", levelStrings[rec.Level], "] ", rec.Message, "\n")
	}
}

func (w ConsoleLogWriter) runWithColor(out io.Writer) {
	var timestr string
	//var timestrAt int64

	for rec := range w {
		//if at := rec.Created.UnixNano() / 1e9; at != timestrAt {
		//	timestr, timestrAt = rec.Created.Format("01/02/06 15:04:05"), at
		//}
		timestr = fmt.Sprintf("%s.%06d", rec.Created.Format("01/02/06 15:04:05"), uint32(float32(rec.Created.Nanosecond())*0.001))
		fmt.Fprintf(out, console.ColorfulText(levelColor[rec.Level], fmt.Sprintf("[%s] [%s] (%s) %s\n", timestr, levelStrings[rec.Level], rec.Source, rec.Message)))
	}
}

// This is the ConsoleLogWriter's output method.  This will block if the output
// buffer is full.
func (w ConsoleLogWriter) LogWrite(rec *LogRecord) {
	w <- rec
}

// Close stops the logger from sending messages to standard output.  Attempts to
// send log messages to this logger after a Close have undefined behavior.
func (w ConsoleLogWriter) Close() {
	close(w)
}
