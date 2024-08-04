// Copyright lowRISC contributors (OpenTitan project).
// Licensed under the Apache License, Version 2.0, see LICENSE for details.
// SPDX-License-Identifier: Apache-2.0

// Package lex provides a basic lexer for parsing commands off of a REPL.
package lex

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLex(t *testing.T) {
	// We use a different type than lex.Token here for:
	// - Convenience.
	// - Not having to test the Position.
	type tok struct {
		Value any
		Text  string
	}

	tests := []struct {
		name, text string
		want       []tok
		wantErr    bool
	}{
		{
			name: "empty",
			text: ``,
			want: []tok{{EOF{}, ``}},
		},
		{
			name: "single word",
			text: `foo`,
			want: []tok{
				{Str("foo"), `foo`},
				{EOF{}, ``},
			},
		},
		{
			name: "multiple words",
			text: ` foo    $bar b_a-z`,
			want: []tok{
				{Str("foo"), `foo`},
				{Var("bar"), `$bar`},
				{Str("b_a-z"), `b_a-z`},
				{EOF{}, ``},
			},
		},
		{
			name: "multi-line",
			text: `foo $bar
             baz`,
			want: []tok{
				{Str("foo"), `foo`},
				{Var("bar"), `$bar`},
				{End('\n'), "\n"},
				{Str("baz"), `baz`},
				{EOF{}, ``},
			},
		},
		{
			name: "semicolon",
			text: `foo $bar; baz`,
			want: []tok{
				{Str("foo"), `foo`},
				{Var("bar"), `$bar`},
				{End(';'), ";"},
				{Str("baz"), `baz`},
				{EOF{}, ``},
			},
		},
		{
			name: "empty quotes",
			text: `foo ""`,
			want: []tok{
				{Str("foo"), `foo`},
				{Str(""), `""`},
				{EOF{}, ``},
			},
		},
		{
			name: "ints",
			text: `my-ints 123 0b1010 0xdeadbeef`,
			want: []tok{
				{Str("my-ints"), `my-ints`},
				{Int(123), `123`},
				{Int(0xa), `0b1010`},
				{Int(0xdeadbeef), `0xdeadbeef`},
				{EOF{}, ``},
			},
		},
		{
			name:    "signed ints",
			text:    `-42`,
			want:    []tok{},
			wantErr: true,
		},
		{
			name: "quotes",
			text: `"foo" "bar baz"`,
			want: []tok{
				{Str("foo"), `"foo"`},
				{Str("bar baz"), `"bar baz"`},
				{EOF{}, ``},
			},
		},
		{
			name: "escapes",
			text: `"\n\t\x00\x01\\\""`,
			want: []tok{
				{Str("\n\t\x00\x01\\\""), `"\n\t\x00\x01\\\""`},
				{EOF{}, ``},
			},
		},
		{
			name: "quotes packed together",
			text: `a"b"$c`,
			want: []tok{
				{Str("a"), `a`},
				{Str("b"), `"b"`},
				{Var("c"), `$c`},
				{EOF{}, ``},
			},
		},
		{
			name: "broken quotes",
			text: `"foo" "bar`,
			want: []tok{
				{Str("foo"), `"foo"`},
			},
			wantErr: true,
		},
		{
			name: "unknown escape",
			text: `"foo" "bar\z"`,
			want: []tok{
				{Str("foo"), `"foo"`},
			},
			wantErr: true,
		},
		{
			name: "unknown rune",
			text: `foo?`,
			want: []tok{
				{Str("foo"), `foo`},
			},
			wantErr: true,
		},
		{
			name: "hex strings",
			text: `h"" h"deadbeef" h "abc"`,
			want: []tok{
				{Str(""), `h""`},
				{Str(string([]byte{0xde, 0xad, 0xbe, 0xef})), `h"deadbeef"`},
				{Str("h"), `h`},
				{Str("abc"), `"abc"`},
				{EOF{}, ``},
			},
		},
		{
			name:    "bad hex string",
			text:    `h"not hex"`,
			want:    []tok{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lex := New(strings.NewReader(tt.text))

			tokens := []tok{}
			var err error
			for {
				var t Token
				t, err = lex.Next()
				if err != nil {
					break
				}
				tokens = append(tokens, tok{t.Value, t.Text})
				if _, done := t.Value.(EOF); done {
					break
				}
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %s", err)
			}
			if tt.wantErr && err == nil {
				t.Errorf("missing expected error")
			}
			if diff := cmp.Diff(tt.want, tokens); diff != "" {
				t.Errorf("unexpected diff (-want +got):\n%s", diff)
			}
		})
	}
}
