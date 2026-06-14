package sensors

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func benchTree(b *testing.B, files int) string {
	b.Helper()
	root := b.TempDir()
	for i := 0; i < files; i++ {
		dir := filepath.Join(root, "pkg", "sub")
		_ = os.MkdirAll(dir, 0o755)
		body := []byte("package x\n\nvar Foo = \"AKIAIOSFODNN7EXAMPLE\"\n")
		_ = os.WriteFile(filepath.Join(dir, "file"+itoa(i)+".go"), body, 0o644)
	}
	return root
}

func BenchmarkForbiddenFiles_100files(b *testing.B) {
	root := benchTree(b, 100)
	rc := RunCtx{Ctx: context.Background(), Root: root}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ForbiddenFilesSensor{}.Run(rc)
	}
}

func BenchmarkSecretsScan_100files(b *testing.B) {
	root := benchTree(b, 100)
	rc := RunCtx{Ctx: context.Background(), Root: root}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SecretsScanSensor{}.Run(rc)
	}
}

func BenchmarkForbiddenCommands_50files(b *testing.B) {
	root := b.TempDir()
	for i := 0; i < 50; i++ {
		_ = os.WriteFile(filepath.Join(root, "s"+itoa(i)+".sh"),
			[]byte("#!/bin/sh\nrm -rf /tmp/x\nchmod -R 777 /tmp/y\n"), 0o755)
	}
	rc := RunCtx{Ctx: context.Background(), Root: root}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ForbiddenCommandsSensor{}.Run(rc)
	}
}
