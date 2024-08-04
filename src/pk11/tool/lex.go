// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package lex provides a basic lexer for parsing commands off of a REPL.
package lex

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"text/scanner"
	"unicode"
)

// Error represents a parsing error returned by Lex.Next()
type Error struct {
	Msg string
	scanner.Position
}

// Error converts an Error into a string.
func (e *Error) Error() string {
	return fmt.Sprintf("error at %s: %s", e.Position, e.Msg)
}

// Lex is a lexer, which converts a stream of runes into a stream of tokens.
type Lex struct {
	r    io.Reader
	scan scanner.Scanner
	err  *Error
	eol  bool
}

// New creates a new Lex over the given io.Reader.
func New(r io.Reader) *Lex {
	l := &Lex{}
	l.r = r
	l.scan.Init(l.r)
	l.scan.Error = func(s *scanner.Scanner, msg string) {
		l.err = &Error{msg, s.Pos()}
	}
	l.scan.Mode = scanner.ScanIdents | scanner.ScanInts | scanner.ScanStrings | scanner.ScanComments | scanner.SkipComments
	l.scan.Whitespace &^= 1 << '\n' // Don't skip newlines!
	l.scan.IsIdentRune = func(ch rune, i int) bool {
		if i == 0 && ch == '$' {
			return true
		}
		if i != 0 && (unicode.IsDigit(ch) || ch == '-') {
			return true
		}
		return ch == '_' || unicode.IsLetter(ch)
	}

	return l
}

// Var represents a $-prefixed variable reference, like $foo.
type Var string

// Str represents a string, either "quoted", h"deadbeef", or as a bare-word.
type Str string

// Int represents an integer.
type Int int64

// End represents a statement-terminator, like ; or \n.
type End rune

// EOF represents an end-of-file.
type EOF struct{}

// Token is a token returned by Lex.Next()
type Token struct {
	// One of Var, Str, Int, End, or EOF.
	Value any
	// The text the token was parsed from.
	Text string
	// The position the token was found at.
	scanner.Position
}

// String stringifies a Token.
func (t Token) String() string {
	return t.Text
}

// StringTokens stringifies an array of Tokens.
func StringTokens(tokens []Token) string {
	var b strings.Builder
	for i, t := range tokens {
		if i != 0 {
			fmt.Fprint(&b, " ")
		}
		fmt.Fprint(&b, t.Text)
	}
	return b.String()
}

// Next advances the underlying reader until a token is parsed or an error
// is encountered.
func (l *Lex) Next() (tok Token, err error) {
	if l.eol {
		l.eol = false
		_ = l.scan.Next()
	}

	// Scan() will skip over the current token, which, if this is something like stdin,
	// will cause it to hang indefinitely; Next() also has this problem.
	// Therefore, we first peek the next rune; if it is a newline, and if we can detect
	// that the reader hasn't hit EOF but isn't ready yet, we return immediately.
	if f, ok := l.r.(*os.File); ok {
		stat, e := f.Stat()
		if e != nil {
			err = e
			return
		}
		if stat.Size() == 0 && l.scan.Peek() == '\n' {
			l.eol = true
			tok = Token{End('\n'), "\n", l.scan.Pos()}
			return
		}
	}

	l.err = nil
	next := l.scan.Scan()
	if l.err != nil {
		err = l.err
		return
	}

	tok.Position = l.scan.Pos()
	switch next {
	case scanner.Ident:
		tok.Text = l.scan.TokenText()
		if tok.Text == "h" && l.scan.Peek() == '"' {
			// This is a hex string; process it as such.
			tok2 := l.scan.Scan()
			if l.err != nil {
				err = l.err
				return
			}
			if tok2 != scanner.String {
				err = fmt.Errorf("expected string constant")
				return
			}
			tok.Text += l.scan.TokenText()
			hexStr := tok.Text[2 : len(tok.Text)-1]

			dec, e := hex.DecodeString(hexStr)
			if e != nil {
				err = e
				return
			}

			tok.Value = Str(dec)
		} else if tok.Text[0] == '$' {
			tok.Value = Var(tok.Text[1:])
		} else {
			tok.Value = Str(tok.Text)
		}
	case scanner.Int:
		tok.Text = l.scan.TokenText()
		var i int64
		i, err = strconv.ParseInt(tok.Text, 0, 64)
		tok.Value = Int(i)
	case scanner.String:
		tok.Text = l.scan.TokenText()
		var unquote string
		unquote, err = strconv.Unquote(tok.Text)
		tok.Value = Str(unquote)
	case ';', '\n':
		tok.Text = string(next)
		tok.Value = End(next)
	case scanner.EOF:
		tok.Value = EOF{}
	default:
		err = fmt.Errorf("unrecognized rune: %v", next)
	}

	return
}
