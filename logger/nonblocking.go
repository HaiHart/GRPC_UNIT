package logger

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

const backlog = 100000

type line struct {
	data  *Fields
	msg   string
	level Level
	time  time.Time
}

type nonBlocking struct {
	lines        chan line
	avoidChannel bool
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// NonBlocking log
var NonBlocking *nonBlocking

func init() {
}

func logOneLine(l line) {
	entry := logrus.NewEntry(logrus.StandardLogger())
	entry.Time = l.time
	if l.data != nil {
		entry = entry.WithFields(logrus.Fields(*l.data))
	}
	entry.Log(logrus.Level(l.level), l.msg)
}

func newNonBlockingLogger() *nonBlocking {
	lines := make(chan line, backlog)
	ctx, cancel := context.WithCancel(context.Background())
	l := &nonBlocking{lines: lines, ctx: ctx, cancel: cancel}
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		for {
			select {
			case oneline, ok := <-lines:
				if !ok {
					logrus.Errorf("non blocking log channel unexpectdly closed. nonBlockingLogger stopped")
					return
				}
				logOneLine(oneline)
			case <-ctx.Done():
				for {
					select {
					case oneline := <-lines:
						logOneLine(oneline)
					default:
						return
					}
				}
			}
		}
	}()
	return l
}

// AvoidChannel should be used in tests only
func (l *nonBlocking) AvoidChannel() {
	l.avoidChannel = true
}

// Logf - main log function
func (l *nonBlocking) Logf(level Level, fields *Fields, format string, args ...interface{}) {
	if !IsLevelEnabled(level) {
		return
	}

	oneLine := line{
		msg:   fmt.Sprintf(format, args),
		data:  fields,
		level: level,
		time:  time.Now(),
	}

	if l.avoidChannel {
		logOneLine(oneLine)
		return
	}

	select {
	case l.lines <- oneLine:
	default:
		logrus.Errorf("Unable to use logger properly")
	}
}

func (l *nonBlocking) Log(level Level, fields *Fields, a ...interface{}) {
	if !IsLevelEnabled(level) {
		return
	}
	oneLine := line{
		msg:   fmt.Sprint(a...),
		data:  fields,
		level: level,
		time:  time.Now(),
	}
	if l.avoidChannel {
		logOneLine(oneLine)
		return
	}

	select {
	case l.lines <- oneLine:
	default:
		logrus.Errorf("Unable to use FastLogger properly")
	}
}

// Exit verifies all log records in the channel are written before exiting
func (l *nonBlocking) Exit() {
	l.AvoidChannel()
	// Flushes all log records
	l.cancel()
	l.wg.Wait()
}
