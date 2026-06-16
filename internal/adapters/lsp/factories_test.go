// SPDX-License-Identifier: MIT

package lsp

import "testing"

func TestNewGoplsConstructor(t *testing.T) {
	c := NewGopls("/tmp")
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewPyrightConstructor(t *testing.T) {
	c := NewPyright("/tmp")
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewBasedPyrightConstructor(t *testing.T) {
	c := NewBasedPyright("/tmp")
	if c == nil {
		t.Fatal("nil client")
	}
}

func TestNewRustAnalyzerConstructor(t *testing.T) {
	c := NewRustAnalyzer("/tmp")
	if c == nil {
		t.Fatal("nil client")
	}
}
