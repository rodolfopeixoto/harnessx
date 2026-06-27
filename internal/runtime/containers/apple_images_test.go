package containers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func writeAppleStub(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "container")
	if err := os.WriteFile(bin, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
	return bin
}

func TestAppleListImagesArrayOutput(t *testing.T) {
	stub := writeAppleStub(t, `#!/bin/sh
cat <<'EOF'
[
  {"id":"sha256:abc","repository":"alpine","tag":"3.20","size":5242880,"created":"2026-06-01T00:00:00Z"},
  {"id":"sha256:def","repository":"python","tag":"3.12","size":104857600}
]
EOF
`)
	imgs, err := appleListImages(context.Background(), stub)
	if err != nil {
		t.Fatalf("listImages: %v", err)
	}
	if len(imgs) != 2 {
		t.Fatalf("want 2 images, got %d: %+v", len(imgs), imgs)
	}
	if imgs[0].Repository != "alpine" || imgs[0].Tag != "3.20" {
		t.Errorf("first image fields wrong: %+v", imgs[0])
	}
	if imgs[1].SizeBytes != 104857600 {
		t.Errorf("size parsed wrong: %+v", imgs[1])
	}
}

func TestAppleListImagesNDJSONOutput(t *testing.T) {
	stub := writeAppleStub(t, `#!/bin/sh
cat <<'EOF'
{"id":"sha256:1","repository":"r1","tag":"latest"}
{"id":"sha256:2","repository":"r2","tag":"v1"}
EOF
`)
	imgs, err := appleListImages(context.Background(), stub)
	if err != nil {
		t.Fatalf("listImages: %v", err)
	}
	if len(imgs) != 2 {
		t.Fatalf("want 2 images, got %d", len(imgs))
	}
}

func TestAppleListImagesEmpty(t *testing.T) {
	stub := writeAppleStub(t, "#!/bin/sh\necho '[]'\n")
	imgs, err := appleListImages(context.Background(), stub)
	if err != nil {
		t.Fatalf("listImages: %v", err)
	}
	if len(imgs) != 0 {
		t.Fatalf("want 0 images, got %d", len(imgs))
	}
}

func TestAppleListImagesError(t *testing.T) {
	stub := writeAppleStub(t, "#!/bin/sh\necho 'boom: registry unavailable' >&2\nexit 2\n")
	_, err := appleListImages(context.Background(), stub)
	if err == nil {
		t.Fatal("expected error on exit 2")
	}
}
