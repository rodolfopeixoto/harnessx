// SPDX-License-Identifier: MIT

package hashing

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSHA256Bytes(t *testing.T) {
	require.Equal(t,
		"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		SHA256Bytes(nil))
	require.Equal(t,
		"2c26b46b68ffc68ff99b453c1d30413413422d706483bfa0f98a5e886266e7ae",
		SHA256Bytes([]byte("foo")))
}

func TestSHA256File_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.txt")
	require.NoError(t, os.WriteFile(p, []byte("foo"), 0o644))
	got, err := SHA256File(p)
	require.NoError(t, err)
	require.Equal(t, SHA256Bytes([]byte("foo")), got)
}

func TestSHA256File_Missing(t *testing.T) {
	_, err := SHA256File(filepath.Join(t.TempDir(), "nope"))
	require.Error(t, err)
}
