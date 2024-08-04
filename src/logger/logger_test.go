// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package logger implements wrapper for standard log package.
//
// Outputs log to console and log file with file rotation.

package logger

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

var (
	tempLogFile    string
	invalidLogFile string
)

func init() {
	if runtime.GOOS == "windows" {
		tempLogFile = filepath.Join(os.TempDir(), "test.log")
		dir, _ := os.Getwd()
		invalidLogFile = filepath.Join(dir, "log", "test.log")
	} else {
		tempLogFile = filepath.Join(os.Getenv("TEST_TMPDIR"), "test.log")
		invalidLogFile = "/test/test.log"
	}
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		name string
		l    LogLevel
		want string
	}{
		{
			name: "ValidLogLevel",
			l:    LogLevelWarn,
			want: "WARN: ",
		},
		{
			name: "InvalidLogLevel",
			l:    10,
			want: "10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.l.String(); got != tt.want {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_rotate(t *testing.T) {
	type args struct {
		l *ModLogger
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	newLog.Info(errors.New("Test info"), "Info message", 123)
	newLog.CreateTime = time.Now().Add(-time.Hour * 24 * 8)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ValidLogPath",
			args: args{
				l: newLog,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := rotate(tt.args.l); (err != nil) != tt.wantErr {
				t.Errorf("rotate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	newLog.CreateTime = time.Now()
	newLog.Info(errors.New("Test info"), "Info message", 456)

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err != nil {
			t.Errorf("cannot close log file error = %v", err)
		}

		files, err := filepath.Glob(tempLogFile + "*")
		if err != nil {
			t.Errorf("cannot create log file pattern error = %v", err)
		}

		for _, f := range files {
			if err := os.Remove(f); err != nil {
				t.Errorf("cannot remove %s file error = %v", f, err)
			}
		}
	}
}

func Test_getPrefix(t *testing.T) {
	type args struct {
		err error
		l   LogLevel
	}

	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPrefix(tt.args.err, tt.args.l); got != tt.want {
				t.Errorf("getPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	type args struct {
		logName  string
		logLevel LogLevel
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ValidLogPath",
			args: args{
				logName:  tempLogFile,
				logLevel: LogLevelInfo,
			},
			wantErr: false,
		},
		{
			name: "EmptyFileName",
			args: args{
				logName:  "",
				logLevel: LogLevelInfo,
			},
			wantErr: false,
		},
		{
			name: "InvalidLogPath",
			args: args{
				logName:  invalidLogFile,
				logLevel: LogLevelInfo,
			},
			wantErr: true,
		},
		{
			name: "InvalidLogLevel",
			args: args{
				logName:  tempLogFile,
				logLevel: 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLogger(tt.args.logName, tt.args.logLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil && got == nil {
				t.Errorf("NewLogger() returned nil ModLogger unexpectedly")
				return
			}

			if got != nil {
				if got.FatalLog == nil || got.ErrorLog == nil ||
					got.WarnLog == nil || got.InfoLog == nil ||
					got.DebugLog == nil || got.TraceLog == nil {
					t.Errorf("NewLogger() all logs are nil, want non-nil")
				}

				if got.LogFile != nil {
					if err := got.LogFile.Close(); err == nil {
						if err := os.Remove(got.LogFile.Name()); err != nil {
							t.Errorf("cannot remove log file error = %v", err)
						}
					}
				}
			}
		})
	}
}

func TestModLogger_SetLogLevel(t *testing.T) {
	type args struct {
		logLevel LogLevel
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name    string
		l       *ModLogger
		args    args
		wantErr bool
	}{
		{
			name: "ValidLevel",
			l:    newLog,
			args: args{
				logLevel: LogLevelDebug,
			},
			wantErr: false,
		},
		{
			name: "InvalidLevel",
			l:    newLog,
			args: args{
				logLevel: 10,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.l.SetLogLevel(tt.args.logLevel); (err != nil) != tt.wantErr {
				t.Errorf("ModLogger.SetLogLevel() error = %v, wantErr %v",
					err, tt.wantErr)
			}
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Fatal(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidFatal",
			l:    newLog,
			args: args{
				err:  errors.New("test fatal"),
				intf: []interface{}{"Fatal message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test fatal"),
				intf: []interface{}{"Fatal message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Fatal(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Panic(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidPanic",
			l:    newLog,
			args: args{
				err:  errors.New("test panic"),
				intf: []interface{}{"Panic message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test panic"),
				intf: []interface{}{"Panic message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Error(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Error(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidError",
			l:    newLog,
			args: args{
				err:  errors.New("test error"),
				intf: []interface{}{"Error message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test error"),
				intf: []interface{}{"Error message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Error(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Warn(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidWarn",
			l:    newLog,
			args: args{
				err:  errors.New("test warn"),
				intf: []interface{}{"Warn message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test warn"),
				intf: []interface{}{"Warn message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Warn(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Info(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidInfo",
			l:    newLog,
			args: args{
				err:  errors.New("test error"),
				intf: []interface{}{"Info message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test error"),
				intf: []interface{}{"Info message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Info(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Debug(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile, LogLevelDebug)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidDebug",
			l:    newLog,
			args: args{
				err:  errors.New("test debug"),
				intf: []interface{}{"Debug message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test debug"),
				intf: []interface{}{"Debug message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Info(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}

func TestModLogger_Trace(t *testing.T) {
	type args struct {
		err  error
		intf []interface{}
	}

	newLog, err := NewLogger(tempLogFile, LogLevelTrace)
	if err != nil {
		t.Errorf("NewLogger() error = %v", err)
		return
	}

	tests := []struct {
		name string
		l    *ModLogger
		args args
	}{
		{
			name: "ValidTrace",
			l:    newLog,
			args: args{
				err:  errors.New("test trace"),
				intf: []interface{}{"Trace message", 123},
			},
		},
		{
			name: "InvalidLog",
			l: &ModLogger{
				LogFile: nil,
			},
			args: args{
				err:  errors.New("test trace"),
				intf: []interface{}{"Trace message", 123},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.l.Info(tt.args.err, tt.args.intf...)
		})
	}

	if newLog.LogFile != nil {
		if err := newLog.LogFile.Close(); err == nil {
			if err := os.Remove(newLog.LogFile.Name()); err != nil {
				t.Errorf("cannot remove log file error = %v", err)
			}
		}
	}
}
