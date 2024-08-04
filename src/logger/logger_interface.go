// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package logger implements wrapper for standard log package.
//
// Outputs log to console and log file with file rotation.

package logger

type Logger interface {
	NewLogger(logName string, logLevel ...LogLevel) (Logger, error)
	DeleteLogger() error
	SetLogLevel(logLevel LogLevel) error
	Fatal(err error, intf ...interface{})
	Panic(err error, intf ...interface{})
	Error(err error, intf ...interface{})
	Warn(err error, intf ...interface{})
	Info(err error, intf ...interface{})
	Debug(err error, intf ...interface{})
	Trace(err error, intf ...interface{})
}
