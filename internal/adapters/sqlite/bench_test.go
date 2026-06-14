package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/ropeixoto/harnessx/internal/domain"
)

func BenchmarkCreateSessionRun(b *testing.B) {
	repo, err := Open(":memory:")
	if err != nil {
		b.Fatal(err)
	}
	defer repo.Close()
	ctx := context.Background()
	now := time.Now().UTC()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s := domain.Session{ID: itoa(i), ProjectPath: "/tmp", Mode: domain.ModeBootstrap, Status: domain.StatusRunning, StartedAt: now}
		if err := repo.CreateSession(ctx, s); err != nil {
			b.Fatal(err)
		}
		r := domain.Run{ID: "r" + itoa(i), SessionID: s.ID, Stage: domain.StageInit, Status: domain.StatusRunning, StartedAt: now}
		if err := repo.CreateRun(ctx, r); err != nil {
			b.Fatal(err)
		}
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}
