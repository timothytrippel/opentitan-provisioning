// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package logger implements wrapper for standard log package.
//
// Outputs log to console and log file with file rotation.

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"
)

const (
	DDMMYYYYhhmmss = "20060102150405"
)

type LogLevel int

const (
	LogLevelFatal LogLevel = iota
	LogLevelPanic
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
	LogLevelTrace
)

type ModLogger struct {
	FatalLog   *log.Logger
	PanicLog   *log.Logger
	ErrorLog   *log.Logger
	WarnLog    *log.Logger
	InfoLog    *log.Logger
	DebugLog   *log.Logger
	TraceLog   *log.Logger
	LogFile    *os.File
	CreateTime time.Time
	LogMutex   sync.Mutex
	RefCount   int
}

var (
	level   LogLevel
	loggers = make(map[string]*ModLogger)
)

func (level LogLevel) String() string {
	switch level {
	case LogLevelFatal:
		return "FATAL:"
	case LogLevelPanic:
		return "PANIC:"
	case LogLevelError:
		return "ERROR:"
	case LogLevelWarn:
		return "WARN: "
	case LogLevelInfo:
		return "INFO: "
	case LogLevelDebug:
		return "DEBUG:"
	case LogLevelTrace:
		return "TRACE:"
	default:
		return fmt.Sprintf("%d", int(level))
	}
}

func rotate(l *ModLogger) error {
	time1 := time.Now()
	time2 := l.CreateTime
	diff := time1.Sub(time2)
	weekTime := time.Hour * 24 * 7

	if diff >= weekTime {
		name := l.LogFile.Name()

		l.LogMutex.Lock()
		defer l.LogMutex.Unlock()

		oldLog := filepath.Join(name + "_" + time1.Format(DDMMYYYYhhmmss))
		oldFile, err := os.Create(oldLog)
		if err != nil {
			return fmt.Errorf("cannot create %s file %w", oldLog, err)
		}
		defer oldFile.Close()

		l.LogFile.Seek(0, 0)

		fileInfo, err := os.Stat(name)
		if err != nil {
			return fmt.Errorf("cannot get log file info %w", err)
		}

		fileSize := fileInfo.Size()
		buf := make([]byte, fileSize)

		_, err = l.LogFile.Read(buf)
		if err != nil && err != io.EOF {
			return fmt.Errorf("cannot read from log file %w", err)
		}

		_, err = oldFile.Write(buf)
		if err != nil {
			return fmt.Errorf("cannot write to log file %w", err)
		}

		err = os.Truncate(name, 0)
		if err != nil {
			return fmt.Errorf("cannot truncate log file %w", err)
		}

		l.CreateTime = time.Now()
	}

	return nil
}

func getPrefix(err error, l LogLevel) string {
	now := time.Now()
	s := fmt.Sprintf("%s %s %s", now.Format(DDMMYYYYhhmmss), l.String(),
		err.Error())

	pc, path, line, ok := runtime.Caller(2)
	if ok {
		details := runtime.FuncForPC(pc)
		_, file := filepath.Split(path)
		s = fmt.Sprintf("%s %s [%s()] [%s] [%d] %s", now.Format(DDMMYYYYhhmmss),
			l.String(), details.Name(), file, line, err.Error())
	}

	return s
}

func NewLogger(logName string, logLevel ...LogLevel) (*ModLogger, error) {
	level = LogLevelInfo

	if len(logLevel) > 0 {
		if logLevel[0] < LogLevelFatal || logLevel[0] > LogLevelTrace {
			return nil,
				fmt.Errorf("invalid log level %d, expected from %d to %d",
					logLevel[0], LogLevelFatal, LogLevelTrace)
		}

		level = logLevel[0]
	}

	var logger *ModLogger
	loggers := make(map[string]*ModLogger)

	if logName == "" {
		createTime := time.Now()
		fatalLog := log.New(os.Stderr, "", 0)
		panicLog := log.New(os.Stderr, "", 0)
		errorLog := log.New(os.Stderr, "", 0)
		warnLog := log.New(os.Stderr, "", 0)
		infoLog := log.New(os.Stderr, "", 0)
		debugLog := log.New(os.Stderr, "", 0)
		traceLog := log.New(os.Stderr, "", 0)

		wrt := os.Stderr
		fatalLog.SetOutput(wrt)
		panicLog.SetOutput(wrt)
		errorLog.SetOutput(wrt)
		warnLog.SetOutput(wrt)
		infoLog.SetOutput(wrt)
		debugLog.SetOutput(wrt)
		traceLog.SetOutput(wrt)

		logger = &ModLogger{
			FatalLog:   fatalLog,
			PanicLog:   panicLog,
			ErrorLog:   errorLog,
			WarnLog:    warnLog,
			InfoLog:    infoLog,
			DebugLog:   debugLog,
			TraceLog:   traceLog,
			CreateTime: createTime,
			LogMutex:   sync.Mutex{},
			RefCount:   0,
		}
	} else {
		var ok bool
		logger, ok = loggers[logName]
		if ok {
			logger.RefCount++
		} else {
			_, err := os.Stat(filepath.Dir(logName))
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("log directory %s does not exist",
					filepath.Dir(logName))
			}

			logFile, err := os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
			if err != nil {
				return nil, fmt.Errorf("cannot create log file %w", err)
			}

			createTime := time.Now()
			fatalLog := log.New(logFile, "", 0)
			panicLog := log.New(logFile, "", 0)
			errorLog := log.New(logFile, "", 0)
			warnLog := log.New(logFile, "", 0)
			infoLog := log.New(logFile, "", 0)
			debugLog := log.New(logFile, "", 0)
			traceLog := log.New(logFile, "", 0)

			wrt := io.MultiWriter(os.Stderr, logFile)
			fatalLog.SetOutput(wrt)
			panicLog.SetOutput(wrt)
			errorLog.SetOutput(wrt)
			warnLog.SetOutput(wrt)
			infoLog.SetOutput(wrt)
			debugLog.SetOutput(wrt)
			traceLog.SetOutput(wrt)

			logger = &ModLogger{
				FatalLog:   fatalLog,
				PanicLog:   panicLog,
				ErrorLog:   errorLog,
				WarnLog:    warnLog,
				InfoLog:    infoLog,
				DebugLog:   debugLog,
				TraceLog:   traceLog,
				LogFile:    logFile,
				CreateTime: createTime,
				LogMutex:   sync.Mutex{},
				RefCount:   1,
			}

			loggers[logName] = logger
		}
	}

	return logger, nil
}

func (l *ModLogger) Close() error {
	if l == nil {
		return fmt.Errorf("non-existing logger")
	}

	if l.LogFile != nil {
		l.RefCount--
		if l.RefCount == 0 {
			l.LogMutex.Lock()
			defer l.LogMutex.Unlock()

			name := l.LogFile.Name()
			err := l.LogFile.Close()
			if err != nil {
				return fmt.Errorf("cannot close log file %w", err)
			}

			info, err := os.Stat(name)
			if err != nil {
				return fmt.Errorf("cannot get log file info %w", err)
			}

			if info.Size() == 0 {
				err := os.Remove(name)
				if err != nil {
					return fmt.Errorf("cannot remove empty log file %w", err)
				}
			}

			l.LogFile = nil
		}
	}

	return nil
}

func (l *ModLogger) SetLogLevel(logLevel LogLevel) error {
	if logLevel < LogLevelFatal || logLevel > LogLevelTrace {
		return fmt.Errorf("invalid log level %d, expected from %d to %d",
			logLevel, LogLevelFatal, LogLevelTrace)
	}

	level = logLevel

	return nil
}

func (l *ModLogger) Fatal(err error, intf ...interface{}) {
	if l == nil || level < LogLevelFatal {
		return
	}

	s := getPrefix(err, LogLevelFatal)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.FatalLog.Println(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Panic(err error, intf ...interface{}) {
	if l == nil || level < LogLevelPanic {
		return
	}

	s := getPrefix(err, LogLevelPanic)

	if l.LogFile == nil {
		panic(s)
	} else {
		l.PanicLog.Panicln(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Error(err error, intf ...interface{}) {
	if l == nil || level < LogLevelError {
		return
	}

	s := getPrefix(err, LogLevelError)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.ErrorLog.Println(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Warn(err error, intf ...interface{}) {
	if l == nil || level < LogLevelWarn {
		return
	}

	s := getPrefix(err, LogLevelWarn)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.WarnLog.Println(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Info(err error, intf ...interface{}) {
	if l == nil || level < LogLevelInfo {
		return
	}

	s := getPrefix(err, LogLevelInfo)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.InfoLog.Println(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Debug(err error, intf ...interface{}) {
	if l == nil || level < LogLevelInfo {
		return
	}

	s := getPrefix(err, LogLevelDebug)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.DebugLog.Println(s, intf)
		rotate(l)
	}
}

func (l *ModLogger) Trace(err error, intf ...interface{}) {
	if l == nil || level < LogLevelInfo {
		return
	}

	s := getPrefix(err, LogLevelTrace)

	if l.LogFile == nil {
		fmt.Fprintln(os.Stderr, s)
	} else {
		l.TraceLog.Println(s, intf)
		rotate(l)
	}
}
