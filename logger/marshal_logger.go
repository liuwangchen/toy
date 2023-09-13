package logger

import (
	"container/list"
	"fmt"
	"os"
	"sync"
	"time"
)

var newLine = []byte{10}

type marshalFunc func(interface{}) ([]byte, error)

// MarshalLogWriter 序列化一个结构写入log文件
type MarshalLogWriter struct {
	inRec chan interface{}
	rec   chan interface{}
	rot   chan bool

	// The opened file
	filename string
	file     *os.File

	// Rotate at linecount
	maxlines          int
	maxlines_curlines int

	// Rotate at size
	maxsize         int
	maxsize_cursize int

	// Rotate daily
	daily          bool
	daily_opendate int

	// Rotate hour
	hour          bool
	hour_opendate int
	// Keep old logfiles (.001, .002, etc)
	rotate  bool
	closeCh chan struct{}
	msgQ    *list.List
	wg      sync.WaitGroup
	marshal marshalFunc
}

// This is the MarshalLogWriter's output method
func (w *MarshalLogWriter) LogWrite(rec interface{}) {
	w.inRec <- rec
}

func (w *MarshalLogWriter) Close() {
	close(w.closeCh)
	w.wg.Wait()

	for moredata := true; moredata; {
		select {
		case rec := <-w.rec:
			w.write(rec)
		default:
			moredata = false
		}
	}

	for e := w.msgQ.Front(); e != nil; e = e.Next() {
		w.write(e.Value)
	}

	if w.file != nil {
		w.file.Close()
	}
}

func (w *MarshalLogWriter) write(rec interface{}) {
	if (w.maxlines > 0 && w.maxlines_curlines >= w.maxlines) ||
		(w.maxsize > 0 && w.maxsize_cursize >= w.maxsize) {
		if err := w.intRotate(false); err != nil {
			fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
			return
		}
	}

	now := time.Now()

	if w.daily && now.Day() != w.daily_opendate {
		if err := w.intRotate(true); err != nil {
			fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
			return
		}
	}

	if w.hour && now.Hour() != w.hour_opendate {
		if err := w.intRotate(true); err != nil {
			fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
			return
		}
	}
	b, err := w.marshal(rec)
	if nil != err {
		fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
		return
	}

	// Perform the write
	n, err := w.file.Write(b)
	if err != nil {
		fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
		return
	}

	if _, err := w.file.Write(newLine); err != nil {
		fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
		return
	}

	// Update the counts
	w.maxlines_curlines++
	w.maxsize_cursize += n + 1
}

// NewMarshalLogWriter creates a new LogWriter which writes to the given file and
func NewMarshalLogWriter(fname string, rotate bool, marshal marshalFunc) *MarshalLogWriter {
	w := &MarshalLogWriter{
		inRec:    make(chan interface{}),
		rec:      make(chan interface{}, 32),
		rot:      make(chan bool),
		filename: fname,
		rotate:   rotate,
		closeCh:  make(chan struct{}),
		msgQ:     list.New(),
		marshal:  marshal,
	}

	// open the file for the first time
	if err := w.intRotate(false); err != nil {
		fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
		return nil
	}

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		var msg interface{}
		var outCh chan interface{}
		for {
			select {
			case info := <-w.inRec:
				w.msgQ.PushBack(info)
				if msg == nil {
					msg = w.msgQ.Remove(w.msgQ.Front())
					outCh = w.rec
				}
			case outCh <- msg:
				if w.msgQ.Len() > 0 {
					msg = w.msgQ.Remove(w.msgQ.Front())
				} else {
					msg = nil
					outCh = nil
				}
			case <-w.closeCh:
				if msg != nil {
					w.msgQ.PushFront(msg)
				}
				return
			}
		}
	}()

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.rot:
				if err := w.intRotate(false); err != nil {
					fmt.Fprintf(os.Stderr, "MarshalLogWriter(%q): %s\n", w.filename, err)
					return
				}
			case rec := <-w.rec:
				w.write(rec)
			case <-w.closeCh:
				return
			}
		}
	}()

	return w
}

// Request that the logs rotate
func (w *MarshalLogWriter) Rotate() {
	w.rot <- true
}

// If this is called in a threaded context, it MUST be synchronized
// last is true when hourRotate is set and hour change
func (w *MarshalLogWriter) intRotate(last bool) error {
	// Close any log file that may be open
	if w.file != nil {
		//fmt.Fprint(w.file, Formatinterface{}(w.trailer, &interface{}{Created: time.Now()}))
		w.file.Close()
	}

	now := time.Now()
	var lastTime time.Time
	if last {
		lastTime = now.Add(time.Duration(-time.Second * 3600))
	} else {
		lastTime = now
	}
	// If we are keeping log files, move it to the next available number
	if w.rotate {
		_, err := os.Lstat(w.filename)
		if err == nil { // file exists
			// Find the next available number
			num := 1
			fname := ""
			for ; err == nil && num <= 999; num++ {
				fname = w.filename + fmt.Sprintf("-%d-%02d-%02d-%02d+", lastTime.Year(), lastTime.Month(), lastTime.Day(), lastTime.Hour()) + fmt.Sprintf("%03d", num)
				_, err = os.Lstat(fname)
			}
			// return error if the last file checked still existed
			if err == nil {
				return fmt.Errorf("Rotate: Cannot find free log number to rename %s\n", w.filename)
			}

			// Rename the file to its newfound home
			err = os.Rename(w.filename, fname)
			if err != nil {
				return fmt.Errorf("Rotate: %s\n", err)
			}
		}
	}

	// Open the log file
	fd, err := os.OpenFile(w.filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return err
	}
	w.file = fd

	// Set the daily open date to the current date
	w.daily_opendate = now.Day()

	w.hour_opendate = now.Hour()

	// initialize rotation values
	w.maxlines_curlines = 0
	w.maxsize_cursize = 0

	return nil
}

// Set rotate at linecount (chainable). Must be called before the first log
// message is written.
func (w *MarshalLogWriter) SetRotateLines(maxlines int) *MarshalLogWriter {
	//fmt.Fprintf(os.Stderr, "MarshalLogWriter.SetRotateLines: %v\n", maxlines)
	w.maxlines = maxlines
	return w
}

// Set rotate at size (chainable). Must be called before the first log message
// is written.
func (w *MarshalLogWriter) SetRotateSize(maxsize int) *MarshalLogWriter {
	//fmt.Fprintf(os.Stderr, "MarshalLogWriter.SetRotateSize: %v\n", maxsize)
	w.maxsize = maxsize
	return w
}

// Set rotate daily (chainable). Must be called before the first log message is
// written.
func (w *MarshalLogWriter) SetRotateDaily(daily bool) *MarshalLogWriter {
	//fmt.Fprintf(os.Stderr, "MarshalLogWriter.SetRotateDaily: %v\n", daily)
	w.daily = daily
	return w
}

func (w *MarshalLogWriter) SetRotateHour(hour bool) *MarshalLogWriter {
	w.hour = hour
	return w
}

// SetRotate changes whether or not the old logs are kept. (chainable) Must be
// called before the first log message is written.  If rotate is false, the
// files are overwritten; otherwise, they are rotated to another file before the
// new log is opened.
func (w *MarshalLogWriter) SetRotate(rotate bool) *MarshalLogWriter {
	//fmt.Fprintf(os.Stderr, "MarshalLogWriter.SetRotate: %v\n", rotate)
	w.rotate = rotate
	return w
}
