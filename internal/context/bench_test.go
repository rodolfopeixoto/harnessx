package context

import (
	stdctx "context"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkBuild_TinyGoProject(b *testing.B) {
	root := b.TempDir()
	_ = os.WriteFile(filepath.Join(root, "go.mod"), []byte("module sample\n\ngo 1.23\n"), 0o644)
	_ = os.WriteFile(filepath.Join(root, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644)

	opts := Options{
		Root: root, Task: "explain main entry",
		Providers: []Provider{TestMapProvider{}, MemoryProvider{}},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		opts.Force = true
		_, _ = Build(stdctx.Background(), opts)
	}
}
