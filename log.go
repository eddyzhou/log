package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	Ldefault   = log.Lshortfile | log.LstdFlags
	TimeLayout = "20060102"
)

type Level uint8

const (
	Ldebug Level = iota
	Linfo
	Lwarn
	Lerror
	Lpanic
	Lfatal
)

var levels = []string{
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"PANIC",
	"FATAL",
}

func ParseLevel(lvl string) (Level, error) {
	switch strings.ToLower(lvl) {
	case "panic":
		return Lpanic, nil
	case "fatal":
		return Lfatal, nil
	case "error":
		return Lerror, nil
	case "warn":
		return Lwarn, nil
	case "info":
		return Linfo, nil
	case "debug":
		return Ldebug, nil
	}

	var l Level
	return l, fmt.Errorf("invalid log Level: %q", lvl)
}

type Logger struct {
	mu           sync.Mutex
	logger       *log.Logger
	level        Level
	shouldRotate bool
	timeSuffix   string
	fileName     string
	fd           *os.File
}

func New(w io.Writer, prefix string, flag int, level Level) *Logger {
	return &Logger{
		logger: log.New(w, prefix, flag),
		level:  level,
	}
}

func NewRotate(fileName string, prefix string, flag int, level Level) *Logger {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		log.Fatal(err)
	}

	return &Logger{
		logger:       log.New(f, prefix, flag),
		level:        level,
		shouldRotate: true,
		fileName:     fileName,
		fd:           f,
		timeSuffix:   time.Now().Format(TimeLayout),
	}
}

var Std = New(os.Stderr, "", Ldefault, Ldebug)

func (l *Logger) rotate() error {
	suffix := time.Now().Format(TimeLayout)

	if suffix != l.timeSuffix {
		if err := l.doRotate(); err != nil {
			return err
		}
	}

	return nil
}

func (l *Logger) doRotate() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	suffix := time.Now().Format(TimeLayout)
	if suffix == l.timeSuffix {
		return nil
	}

	if err := l.fd.Close(); err != nil {
		return err
	}

	lastName := l.fileName + "." + l.timeSuffix
	if err := os.Rename(l.fileName, lastName); err != nil {
		return err
	}

	f, err := os.OpenFile(l.fileName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		return err
	}

	l.logger.SetOutput(f)
	l.fd = f
	l.timeSuffix = suffix

	return nil
}

// When file is opened with appending mode, it's safe to
// write concurrently to a file (within 4k message on Linux).
func (l *Logger) log(lvl Level, v ...interface{}) {
	if lvl < l.level {
		return
	}

	if l.shouldRotate {
		if err := l.rotate(); err != nil {
			log.Fatal("log rotate failed.", err)
		}
	}

	v1 := make([]interface{}, len(v)+1)
	v1[0] = levels[lvl] + " "
	copy(v1[1:], v)

	l.logger.Output(4, fmt.Sprint(v1...))
}

func (l *Logger) logln(lvl Level, v ...interface{}) {
	if lvl < l.level {
		return
	}

	if l.shouldRotate {
		if err := l.rotate(); err != nil {
			log.Fatal("log rotate failed.", err)
		}
	}

	v1 := make([]interface{}, len(v)+1)
	v1[0] = levels[lvl] + " "
	copy(v1[1:], v)

	l.logger.Output(4, fmt.Sprintln(v1...))
}

func (l *Logger) logf(lvl Level, format string, v ...interface{}) {
	if lvl < l.level {
		return
	}

	if l.shouldRotate {
		if err := l.rotate(); err != nil {
			log.Fatal("log rotate failed.", err)
		}
	}

	s := levels[lvl] + " " + fmt.Sprintf(format, v...)
	l.logger.Output(4, s)
}

func (l *Logger) Fatal(v ...interface{}) {
	l.log(Lfatal, v...)
	os.Exit(-1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logf(Lfatal, format, v...)
	os.Exit(-1)
}

func (l *Logger) Fatalln(v ...interface{}) {
	l.logln(Lfatal, v...)
	os.Exit(-1)
}

func (l *Logger) Panic(v ...interface{}) {
	l.log(Lpanic, v...)
	panic(fmt.Sprint(v...))
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	l.logf(Lpanic, format, v...)
	panic(fmt.Sprintf(format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.log(Lerror, v...)
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logf(Lerror, format, v...)
}

func (l *Logger) Warn(v ...interface{}) {
	l.log(Lwarn, v...)
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logf(Lwarn, format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	l.log(Ldebug, v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logf(Ldebug, format, v...)
}

func (l *Logger) Info(v ...interface{}) {
	l.log(Linfo, v...)
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.logf(Linfo, format, v...)
}

func (l *Logger) Print(v ...interface{}) {
	l.log(Linfo, v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.logf(Linfo, format, v...)
}

func (l *Logger) Println(v ...interface{}) {
	l.logln(Linfo, v...)
}

func Info(v ...interface{}) {
	Std.Info(v...)
}

func Infof(format string, v ...interface{}) {
	Std.Infof(format, v...)
}

func Debug(v ...interface{}) {
	Std.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	Std.Debugf(format, v...)
}

func Warn(v ...interface{}) {
	Std.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	Std.Warnf(format, v...)
}

func Error(v ...interface{}) {
	Std.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	Std.Errorf(format, v...)
}

func Panic(v ...interface{}) {
	Std.Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	Std.Panicf(format, v...)
}

func Fatal(v ...interface{}) {
	Std.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	Std.Fatalf(format, v...)
}

func Fatalln(v ...interface{}) {
	Std.Fatalln(v...)
}

func Print(v ...interface{}) {
	Std.Print(v...)
}

func Printf(format string, v ...interface{}) {
	Std.Printf(format, v...)
}

func Println(v ...interface{}) {
	Std.Println(v...)
}
